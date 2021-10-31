package main

import (
	"fmt"
	"net/url"
	"sort"

	"github.com/prometheus/common/log"
)

type VastAiRawOffer map[string]interface{}
type VastAiRawOffers []VastAiRawOffer

func getRawOffersFromApi(result *VastAiApiResults) error {
	var verified, unverified struct {
		Offers VastAiRawOffers `json:"offers"`
	}
	if err := vastApiCall(&verified, "bundles", url.Values{
		"q": {`{"external":{"eq":"false"},"verified":{"eq":"true"},"type":"on-demand","disable_bundling":true}`},
	}); err != nil {
		return err
	}
	if err := vastApiCall(&unverified, "bundles", url.Values{
		"q": {`{"external":{"eq":"false"},"verified":{"eq":"false"},"type":"on-demand","disable_bundling":true}`},
	}); err != nil {
		return err
	}
	result.offersVerified = &verified.Offers
	result.offersUnverified = &unverified.Offers
	return nil
}

func mergeRawOffers(verified VastAiRawOffers, unverified VastAiRawOffers) VastAiRawOffers {
	result := VastAiRawOffers{}
	for _, offer := range verified {
		offer["verified"] = true
		result = append(result, offer)
	}
	for _, offer := range unverified {
		offer["verified"] = false
		result = append(result, offer)
	}
	for _, offer := range result {
		// remove useless fields
		delete(offer, "external")
		delete(offer, "webpage")
		delete(offer, "logo")
		delete(offer, "pending_count")
		delete(offer, "inet_down_billed")
		delete(offer, "inet_up_billed")
		delete(offer, "storage_total_cost")
		delete(offer, "dph_total")
		delete(offer, "rented")
		delete(offer, "is_bid")
	}
	return result
}

func (offers VastAiRawOffers) filter(filter func(VastAiRawOffer) bool) VastAiRawOffers {
	return offers.filter2(filter, nil)
}

func (offers VastAiRawOffers) filter2(filter func(VastAiRawOffer) bool, postProcess func(VastAiRawOffer) VastAiRawOffer) VastAiRawOffers {
	result := VastAiRawOffers{}
	for _, offer := range offers {
		if filter(offer) {
			if postProcess != nil {
				result = append(result, postProcess(offer))
			} else {
				result = append(result, offer)
			}
		}
	}
	return result
}

func (offers VastAiRawOffers) validate() VastAiRawOffers {
	result := offers.filter(func(offer VastAiRawOffer) bool {
		// check if required fields are ok and have a correct type
		_, ok1 := offer["machine_id"].(float64)
		_, ok2 := offer["gpu_name"].(string)
		_, ok3 := offer["num_gpus"].(float64)
		_, ok4 := offer["dph_base"].(float64)
		_, ok5 := offer["rentable"].(bool)
		if ok1 && ok2 && ok3 && ok4 && ok5 {
			return true
		}
		log.Warnln(fmt.Sprintf("Offer is missing required fields: %v", offer))
		return false
	})

	// also log offers with gpu_frac=null (this happens for whatever reason)
	for machineId, offers := range offers.groupByMachineId() {
		bad := false
		for _, offer := range offers {
			if _, ok := offer["gpu_frac"].(float64); !ok {
				bad = true
			}
		}
		if bad {
			log.Warnln(fmt.Sprintf("Offer list inconsistency: machine %d has offers with gpu_frac=null", machineId))
		}
	}

	return result
}

func (offers VastAiRawOffers) groupByMachineId() map[int]VastAiRawOffers {
	grouped := make(map[int]VastAiRawOffers)
	for _, offer := range offers {
		machineId := offer.machineId()
		grouped[machineId] = append(grouped[machineId], offer)
	}
	return grouped
}

func (offers VastAiRawOffers) filterWholeMachines() VastAiRawOffers {
	result := VastAiRawOffers{}

	for machineId, offers := range offers.groupByMachineId() {
		// for each machine:
		// - find out smallest and largest chunk size
		minChunkSize := 10000
		maxChunkSize := 0
		for _, offer := range offers {
			numGpus := offer.numGpus()
			if numGpus < minChunkSize {
				minChunkSize = numGpus
			}
			if numGpus > maxChunkSize {
				maxChunkSize = numGpus
			}
		}

		// - sum gpu numbers over offers smallest chunk offers
		totalGpus := 0
		usedGpus := 0
		for _, offer := range offers {
			numGpus := offer.numGpus()
			if numGpus == minChunkSize {
				totalGpus += numGpus
				if !offer.rentable() {
					usedGpus += numGpus
				}
			}
		}

		// - find whole machine offer
		var wholeOffers []VastAiRawOffer
		for _, offer := range offers {
			if offer.numGpus() == maxChunkSize {
				wholeOffers = append(wholeOffers, offer)
			}
		}

		// - validate: there must be exactly one whole machine offer, and smallest chunks must sum up to 1 largest chunk
		if len(wholeOffers) != 1 || wholeOffers[0].numGpus() != totalGpus {
			// collect list of chunks log message
			chunks := make([]int, 0, len(offers))
			for _, offer := range offers {
				chunks = append(chunks, offer.numGpus())
			}
			sort.Ints(chunks)

			log.Warnln(fmt.Sprintf("Offer list inconsistency: machine %d has invalid chunk split %v",
				machineId, chunks))
			continue
		}

		// - produce modified offer record with added num_gpus_rented and removed gpu_frac etc
		newOffer := VastAiRawOffer{
			"num_gpus_rented": usedGpus,
			"min_chunk":       minChunkSize,
		}
		for k, v := range wholeOffers[0] {
			if k != "gpu_frac" && k != "rentable" && k != "bundle_id" && k != "cpu_cores_effective" {
				newOffer[k] = v
			}
		}
		result = append(result, newOffer)
	}

	return result
}

func (offer VastAiRawOffer) numGpus() int {
	return int(offer["num_gpus"].(float64))
}

func (offer VastAiRawOffer) numGpusRented() int {
	return offer["num_gpus_rented"].(int)
}

func (offer VastAiRawOffer) pricePerGpu() int { // in cents
	return int(offer["dph_base"].(float64) / offer["num_gpus"].(float64) * 100)
}

func (offer VastAiRawOffer) machineId() int {
	return int(offer["machine_id"].(float64))
}

func (offer VastAiRawOffer) gpuName() string {
	return offer["gpu_name"].(string)
}

func (offer VastAiRawOffer) verified() bool {
	return offer["verified"].(bool)
}

func (offer VastAiRawOffer) rentable() bool {
	return offer["rentable"].(bool)
}
