package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
    gpuOccupancyGauge = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "vastai_machine_gpu_occupancy",
            Help: "GPU occupancy state per machine and GPU number.",
        },
        []string{"machine_id", "gpu"},
    )
)

type machineEarningsAPI struct {
	Summary struct {
		TotalGpu  float64 `json:"total_gpu"`
		TotalStor float64 `json:"total_stor"`
		TotalBwu  float64 `json:"total_bwu"`
		TotalBwd  float64 `json:"total_bwd"`
	} `json:"summary"`
	Current struct {
		Balance    float64 `json:"balance"`
		ServiceFee float64 `json:"service_fee"`
		Total      float64 `json:"total"`
		Credit     float64 `json:"credit"`
	} `json:"current"`
	PerMachine []struct {
		MachineID int     `json:"machine_id"`
		GpuEarn   float64 `json:"gpu_earn"`
		StoEarn   float64 `json:"sto_earn"`
		BwuEarn   float64 `json:"bwu_earn"`
		BwdEarn   float64 `json:"bwd_earn"`
	} `json:"per_machine"`
	PerDay []struct {
		Day     int     `json:"day"`
		GpuEarn float64 `json:"gpu_earn"`
		StoEarn float64 `json:"sto_earn"`
		BwuEarn float64 `json:"bwu_earn"`
		BwdEarn float64 `json:"bwd_earn"`
	} `json:"per_day"`
}

type MachinesAPI struct {
	Machines []struct {
		MachineID                     int         `json:"machine_id"`
		Hostname                      string      `json:"hostname"`
		Timeout                       float64     `json:"timeout"`
		NumGpus                       int         `json:"num_gpus"`
		TotalFlops                    float64     `json:"total_flops"`
		GpuName                       string      `json:"gpu_name"`
		GpuRAM                        int         `json:"gpu_ram"`
		GpuMaxCurTemp                 float64     `json:"gpu_max_cur_temp"`
		GpuLanes                      int         `json:"gpu_lanes"`
		GpuMemBw                      float64     `json:"gpu_mem_bw"`
		BwNvlink                      float64     `json:"bw_nvlink"`
		PcieBw                        float64     `json:"pcie_bw"`
		PciGen                        float64     `json:"pci_gen"`
		CPUName                       string      `json:"cpu_name"`
		CPURAM                        int         `json:"cpu_ram"`
		CPUCores                      int         `json:"cpu_cores"`
		Listed                        bool        `json:"listed"`
		CreditDiscountMax             float64     `json:"credit_discount_max"`
		ListedMinGpuCount             int         `json:"listed_min_gpu_count"`
		ListedGpuCost                 float64     `json:"listed_gpu_cost"`
		ListedStorageCost             float64     `json:"listed_storage_cost"`
		ListedInetUpCost              float64     `json:"listed_inet_up_cost"`
		ListedInetDownCost            float64     `json:"listed_inet_down_cost"`
		MinBidPrice                   float64     `json:"min_bid_price"`
		GpuOccupancy                  string      `json:"gpu_occupancy"`
		BidGpuCost                    interface{} `json:"bid_gpu_cost"`
		DiskSpace                     int         `json:"disk_space"`
		MaxDiskSpace                  int         `json:"max_disk_space"`
		AllocDiskSpace                int         `json:"alloc_disk_space"`
		AvailDiskSpace                int         `json:"avail_disk_space"`
		DiskName                      string      `json:"disk_name"`
		DiskBw                        float64     `json:"disk_bw"`
		InetUp                        float64     `json:"inet_up"`
		InetDown                      float64     `json:"inet_down"`
		EarnHour                      float64     `json:"earn_hour"`
		EarnDay                       float64     `json:"earn_day"`
		Verification                  string      `json:"verification"`
		ErrorDescription              interface{} `json:"error_description"`
		CurrentRentalsRunning         int         `json:"current_rentals_running"`
		CurrentRentalsRunningOnDemand int         `json:"current_rentals_running_on_demand"`
		CurrentRentalsResident        int         `json:"current_rentals_resident"`
		CurrentRentalsOnDemand        int         `json:"current_rentals_on_demand"`
		Reliability2                  float64     `json:"reliability2"`
		DirectPortCount               int         `json:"direct_port_count"`
	} `json:"machines"`
}

type VastCollector struct {
	apiKey  string
	metrics map[string]*prometheus.Desc
}

func NewVastCollector(apiKey string) *VastCollector {
	return &VastCollector{
		apiKey: apiKey,
		metrics: map[string]*prometheus.Desc{
			"total_gpu_summary": prometheus.NewDesc(
				"vastai_summary_total_gpu",
				"Total GPU earnings in summary",
				nil, nil,
			),
			"total_stor_summary": prometheus.NewDesc(
				"vastai_summary_total_stor",
				"Total storage earnings in summary",
				nil, nil,
			),
			"total_bwu_summary": prometheus.NewDesc(
				"vastai_summary_total_bwu",
				"Total bandwidth upload earnings in summary",
				nil, nil,
			),
			"total_bwd_summary": prometheus.NewDesc(
				"vastai_summary_total_bwd",
				"Total bandwidth download earnings in summary",
				nil, nil,
			),
			"current_balance": prometheus.NewDesc(
				"vastai_current_balance",
				"Current balance",
				nil, nil,
			),
			"current_service_fee": prometheus.NewDesc(
				"vastai_current_service_fee",
				"Current service fee",
				nil, nil,
			),
			"current_total": prometheus.NewDesc(
				"vastai_current_total",
				"Current total",
				nil, nil,
			),
			"current_credit": prometheus.NewDesc(
				"vastai_current_credit",
				"Current credit",
				nil, nil,
			),
			"per_machine_gpu_earn": prometheus.NewDesc(
				"vastai_per_machine_gpu_earn",
				"GPU earnings per machine",
				[]string{"machine_id"}, nil,
			),
			"per_machine_sto_earn": prometheus.NewDesc(
				"vastai_per_machine_sto_earn",
				"Storage earnings per machine",
				[]string{"machine_id"}, nil,
			),
			"per_machine_bwu_earn": prometheus.NewDesc(
				"vastai_per_machine_bwu_earn",
				"Bandwidth upload earnings per machine",
				[]string{"machine_id"}, nil,
			),
			"per_machine_bwd_earn": prometheus.NewDesc(
				"vastai_per_machine_bwd_earn",
				"Bandwidth download earnings per machine",
				[]string{"machine_id"}, nil,
			),
			"per_day_gpu_earn": prometheus.NewDesc(
				"vastai_per_day_gpu_earn",
				"GPU earnings per day",
				[]string{"day"}, nil,
			),
			"per_day_sto_earn": prometheus.NewDesc(
				"vastai_per_day_sto_earn",
				"Storage earnings per day",
				[]string{"day"}, nil,
			),
			"per_day_bwu_earn": prometheus.NewDesc(
				"vastai_per_day_bwu_earn",
				"Bandwidth upload earnings per day",
				[]string{"day"}, nil,
			),
			"per_day_bwd_earn": prometheus.NewDesc(
				"vastai_per_day_bwd_earn",
				"Bandwidth download earnings per day",
				[]string{"day"}, nil,
			),			
			"machine_id": prometheus.NewDesc(
				"vastai_machine_id",
				"Machine ID",
				nil, nil,
			),
			"machine_timeout": prometheus.NewDesc(
				"vastai_machine_timeout",
				"Machine timeout",
				nil, nil,
			),
			"machine_num_gpus": prometheus.NewDesc(
				"vastai_machine_num_gpus",
				"Number of GPUs in the machine",
				nil, nil,
			),
			"machine_gpu_name": prometheus.NewDesc(
			"vastai_machine_gpu_name",
			"Machine GPU Name",
			[]string{"machine_id", "hostname"}, nil,
			),
			"machine_total_flops": prometheus.NewDesc(
				"vastai_machine_total_flops",
				"Machine total FLOPS",
				nil, nil,
			),
			"machine_Listed": prometheus.NewDesc(
				"vastai_machine_Listed",
				"Machine Listed",
				[]string{"machine_id", "hostname"}, nil,
			),
			"machine_Verification": prometheus.NewDesc(
				"vastai_machine_Verification",
				"Machine Verification",
				[]string{"machine_id", "hostname"}, nil,
			),
			"machine_Reliability": prometheus.NewDesc(
				"vastai_machine_Reliability",
				"Machine Reliability",
				[]string{"machine_id", "hostname"}, nil,
			),
			"machine_InetUp": prometheus.NewDesc(
				"vastai_machine_InetUp",
				"Machine Inet Up",
				[]string{"machine_id", "hostname"}, nil,
			),

			"machine_InetDown": prometheus.NewDesc(
				"vastai_machine_InetDown",
				"Machine Inet Down",
				[]string{"machine_id", "hostname"}, nil,
			),
			// New metrics
			"machine_hostname": prometheus.NewDesc(
				"vastai_machine_hostname",
				"Machine Hostname",
				[]string{"machine_id", "hostname"}, nil,
			),
			"machine_current_rentals_running": prometheus.NewDesc(
				"vastai_machine_current_rentals_running",
				"Current rentals running on machine",
				[]string{"machine_id", "hostname"}, nil,
			),
			"machine_current_rentals_running_on_demand": prometheus.NewDesc(
				"vastai_machine_current_rentals_running_on_demand",
				"Current rentals running on demand on machine",
				[]string{"machine_id", "hostname"}, nil,
			),
			"machine_current_rentals_resident": prometheus.NewDesc(
				"vastai_machine_current_rentals_resident",
				"Current resident rentals on machine",
				[]string{"machine_id", "hostname"}, nil,
			),
			"machine_current_rentals_on_demand": prometheus.NewDesc(
				"vastai_machine_current_rentals_on_demand",
				"Current on-demand rentals on machine",
				[]string{"machine_id", "hostname"}, nil,
			),
			"machine_max_disk_space": prometheus.NewDesc(
				"vastai_machine_max_disk_space",
				"Maximum disk space on machine",
				[]string{"machine_id", "hostname"}, nil,
			),
			"machine_alloc_disk_space": prometheus.NewDesc(
				"vastai_machine_alloc_disk_space",
				"Allocated disk space on machine",
				[]string{"machine_id", "hostname"}, nil,
			),
			"machine_avail_disk_space": prometheus.NewDesc(
				"vastai_machine_avail_disk_space",
				"Available disk space on machine",
				[]string{"machine_id", "hostname"}, nil,
			),
			"gpu_rented_on_demand": prometheus.NewDesc(
				"vastai_machine_gpu_rented_on_demand",
				"Number of GPUs rented on-demand",
				[]string{"machine_id", "hostname"}, nil,
			),
			"gpu_rented_bid_demand": prometheus.NewDesc(
				"vastai_machine_gpu_rented_bid_demand",
				"Number of GPUs rented bid-demand",
				[]string{"machine_id", "hostname"}, nil,
			),
			"gpu_idle": prometheus.NewDesc(
				"vastai_machine_gpu_idle",
				"Number of GPUs idle",
				[]string{"machine_id", "hostname"}, nil,
			),	
			"machine_earn_hour": prometheus.NewDesc(
				"vastai_machine_earn_hour",
				"Machine earn hour",
				[]string{"machine_id", "hostname"}, nil,
			),	
			"machine_ErrorDescription": prometheus.NewDesc(
				"vastai_machine_ErrorDescription",
				"Machine Error Description",
				[]string{"machine_id", "hostname", "error_description"}, nil,
			),
		},
	}
}

func (c *VastCollector) fetchMachineEarnings(ch chan<- prometheus.Metric) {
	earningsURL := fmt.Sprintf("https://console.vast.ai/api/v0/users/me/machine-earnings?api_key=%s", c.apiKey)
	req, err := http.NewRequest("GET", earningsURL, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %s", err)
		return
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to make request: %s", err)
		return
	}
	defer resp.Body.Close()

	var earningsData machineEarningsAPI
	err = json.NewDecoder(resp.Body).Decode(&earningsData)
	if err != nil {
		log.Fatalf("Failed to decode JSON response: %s", err)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.metrics["total_gpu_summary"], prometheus.GaugeValue, earningsData.Summary.TotalGpu)
	ch <- prometheus.MustNewConstMetric(c.metrics["total_stor_summary"], prometheus.GaugeValue, earningsData.Summary.TotalStor)
	ch <- prometheus.MustNewConstMetric(c.metrics["total_bwu_summary"], prometheus.GaugeValue, earningsData.Summary.TotalBwu)
	ch <- prometheus.MustNewConstMetric(c.metrics["total_bwd_summary"], prometheus.GaugeValue, earningsData.Summary.TotalBwd)
	ch <- prometheus.MustNewConstMetric(c.metrics["current_balance"], prometheus.GaugeValue, earningsData.Current.Balance)
	ch <- prometheus.MustNewConstMetric(c.metrics["current_service_fee"], prometheus.GaugeValue, earningsData.Current.ServiceFee)
	ch <- prometheus.MustNewConstMetric(c.metrics["current_total"], prometheus.GaugeValue, earningsData.Current.Total)
	ch <- prometheus.MustNewConstMetric(c.metrics["current_credit"], prometheus.GaugeValue, earningsData.Current.Credit)

	for _, machine := range earningsData.PerMachine {
		ch <- prometheus.MustNewConstMetric(c.metrics["per_machine_gpu_earn"], prometheus.GaugeValue, machine.GpuEarn, strconv.Itoa(machine.MachineID))
		ch <- prometheus.MustNewConstMetric(c.metrics["per_machine_sto_earn"], prometheus.GaugeValue, machine.StoEarn, strconv.Itoa(machine.MachineID))
		ch <- prometheus.MustNewConstMetric(c.metrics["per_machine_bwu_earn"], prometheus.GaugeValue, machine.BwuEarn, strconv.Itoa(machine.MachineID))
		ch <- prometheus.MustNewConstMetric(c.metrics["per_machine_bwd_earn"], prometheus.GaugeValue, machine.BwdEarn, strconv.Itoa(machine.MachineID))
	}

	for _, day := range earningsData.PerDay {
		ch <- prometheus.MustNewConstMetric(c.metrics["per_day_gpu_earn"], prometheus.GaugeValue, day.GpuEarn, strconv.Itoa(day.Day))
		ch <- prometheus.MustNewConstMetric(c.metrics["per_day_sto_earn"], prometheus.GaugeValue, day.StoEarn, strconv.Itoa(day.Day))
		ch <- prometheus.MustNewConstMetric(c.metrics["per_day_bwu_earn"], prometheus.GaugeValue, day.BwuEarn, strconv.Itoa(day.Day))
		ch <- prometheus.MustNewConstMetric(c.metrics["per_day_bwd_earn"], prometheus.GaugeValue, day.BwdEarn, strconv.Itoa(day.Day))
	}
}

func parseGpuOccupancy(occupancy string, machineID string, hostname string, ch chan<- prometheus.Metric) {
    // Remove spaces from the occupancy string
    occupancyNoSpaces := strings.ReplaceAll(occupancy, " ", "")
    
    for i, char := range occupancyNoSpaces {
        state := 0
        switch char {
        case 'R':
            state = 3
        case 'D':
            state = 2
        case 'I':
            state = 1
        case 'x':
            state = 0
        }
        // Emitting the GPU occupancy metric with corrected index
        ch <- prometheus.MustNewConstMetric(
            prometheus.NewDesc("vastai_machine_gpu_occupancy", "GPU occupancy state per machine and GPU number.", []string{"machine_id", "Hostname", "gpu"}, nil),
            prometheus.GaugeValue,
            float64(state),
            machineID,
			hostname,
            strconv.Itoa(i), // i now correctly represents the GPU index
        )
    }
}


func (c *VastCollector) fetchMachines(ch chan<- prometheus.Metric) {
	machinesURL := fmt.Sprintf("https://console.vast.ai/api/v0/machines/?api_key=%s", c.apiKey)
	req, err := http.NewRequest("GET", machinesURL, nil)
	if err != nil {
		log.Fatalf("Failed to create request: %s", err)
		return
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to make request: %s", err)
		return
	}
	defer resp.Body.Close()

	var machinesAPI MachinesAPI
	err = json.NewDecoder(resp.Body).Decode(&machinesAPI)
	if err != nil {
		log.Fatalf("Failed to decode JSON response: %s", err)
		return
	}

	for _, machine := range machinesAPI.Machines {
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("vast_machine_id", "Machine ID", []string{"machine_id", "hostname"}, nil),
			prometheus.GaugeValue,
			float64(machine.MachineID),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("vast_machine_timeout", "Machine timeout", []string{"machine_id", "hostname"}, nil),
			prometheus.GaugeValue,
			float64(machine.Timeout),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("vast_machine_num_gpus", "Number of GPUs in the machine", []string{"machine_id", "hostname"}, nil),
			prometheus.GaugeValue,
			float64(machine.NumGpus),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("vast_machine_gpu_name", "Type and total number of GPUs in the machine", []string{"machine_id", "gpu_name", "hostname"}, nil),
			prometheus.GaugeValue,
			float64(machine.NumGpus),
			strconv.Itoa(machine.MachineID),
			machine.GpuName,
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("vast_machine_total_flops", "Machine total FLOPS", []string{"machine_id", "hostname"}, nil),
			prometheus.GaugeValue,
			machine.TotalFlops,
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		var listedValue float64
		if machine.Listed {
			listedValue = 1.0
		} else {
			listedValue = 0.0
		}	
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("vast_machine_Listed", "Machine Listed", []string{"machine_id", "hostname"}, nil),
			prometheus.GaugeValue,
			listedValue,
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)	
		var verificationValue float64
		if machine.Verification == "verified" {
			verificationValue = 1.0
		} else {
			verificationValue = 0.0
		}	
		
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("vast_machine_Verification", "Machine Verification", []string{"machine_id", "hostname"}, nil),
			prometheus.GaugeValue,
			verificationValue,
			strconv.Itoa(machine.MachineID),
			machine.Hostname,			
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("vast_machine_Reliability", "Machine Reliability", []string{"machine_id", "hostname"}, nil),
			prometheus.GaugeValue,
			machine.Reliability2,
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("vastai_machine_InetUp", "Machine Inet Up", []string{"machine_id", "hostname"}, nil),
			prometheus.GaugeValue,
			machine.InetUp,
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("vast_machine_InetDown", "Machine Inet Down", []string{"machine_id", "hostname"}, nil),
			prometheus.GaugeValue,
			machine.InetDown,
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("vast_machine_hostname", "Machine Hostname", []string{"machine_id", "hostname"}, nil),
			prometheus.GaugeValue,
			1.0,
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		
		ch <- prometheus.MustNewConstMetric(
			c.metrics["machine_current_rentals_running"], 
			prometheus.GaugeValue,
			float64(machine.CurrentRentalsRunning),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["machine_current_rentals_running_on_demand"],
			prometheus.GaugeValue,
			float64(machine.CurrentRentalsRunningOnDemand),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["machine_current_rentals_resident"],
			prometheus.GaugeValue,
			float64(machine.CurrentRentalsResident),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["machine_current_rentals_on_demand"],
			prometheus.GaugeValue,
			float64(machine.CurrentRentalsOnDemand),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["machine_max_disk_space"],
			prometheus.GaugeValue,
			float64(machine.MaxDiskSpace),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["machine_alloc_disk_space"],
			prometheus.GaugeValue,
			float64(machine.AllocDiskSpace),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["machine_avail_disk_space"],
			prometheus.GaugeValue,
			float64(machine.AvailDiskSpace),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)

		gpuRentedOnDemand := strings.Count(machine.GpuOccupancy, "D")
		gpuRentedReserved := strings.Count(machine.GpuOccupancy, "R")
		gpuRentedBidDemand := strings.Count(machine.GpuOccupancy, "I")
		gpuIdle := strings.Count(machine.GpuOccupancy, "x")
		
		ch <- prometheus.MustNewConstMetric(
			c.metrics["gpu_rented_on_demand"],
			prometheus.GaugeValue,
			float64(gpuRentedOnDemand),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["gpu_rented_on_reserved"],
			prometheus.GaugeValue,
			float64(gpuRentedReserved),
			strconv.Itoa(machine.MachineID),
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["gpu_rented_bid_demand"],
			prometheus.GaugeValue,
			float64(gpuRentedBidDemand),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		ch <- prometheus.MustNewConstMetric(
			c.metrics["gpu_idle"],
			prometheus.GaugeValue,
			float64(gpuIdle),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)	

		gpuOccupancy := machine.GpuOccupancy // Ensure this field exists and is correctly named
        machineID := strconv.Itoa(machine.MachineID) // Convert machine ID to string

        parseGpuOccupancy(gpuOccupancy, machineID, machine.Hostname, ch)

		ch <- prometheus.MustNewConstMetric(
			c.metrics["machine_earn_hour"],
			prometheus.GaugeValue,
			float64(machine.EarnHour),
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
		)
		var errorDescription string
		var errorValue float64
		if machine.ErrorDescription == nil {
			errorDescription = ""
			errorValue = 1.0 // No error
		} else {
			errorDescription = fmt.Sprintf("%v", machine.ErrorDescription)
			if errorDescription != "" {
				errorValue = 10.0 // There is an error
			} else {
				errorValue = 1.0 // No error
			}
		}

		// Send ErrorDescription to Prometheus
		ch <- prometheus.MustNewConstMetric(
			c.metrics["machine_ErrorDescription"],
			prometheus.GaugeValue,
			errorValue,
			strconv.Itoa(machine.MachineID),
			machine.Hostname,
			errorDescription,
		)
	}
}




func (c *VastCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric
	}
}

func (c *VastCollector) Collect(ch chan<- prometheus.Metric) {
	c.fetchMachineEarnings(ch)
	c.fetchMachines(ch)
	// Call other fetch methods as you add them
}
