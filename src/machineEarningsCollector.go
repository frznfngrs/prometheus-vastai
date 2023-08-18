package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type machineEarningsCollector struct {
	apiKey          string
	currentBalance   prometheus.Gauge
	currentServiceFee prometheus.Gauge
	totalGPU         prometheus.Gauge
}

// NewMachineEarningsCollector initializes the collector with the required metrics and the API key.
func NewMachineEarningsCollector(apiKey string) *machineEarningsCollector {
	return &machineEarningsCollector{
		apiKey:          apiKey,
		currentBalance:   prometheus.NewGauge(prometheus.GaugeOpts{Name: "vastai_machine_earnings_current_balance"}),
		currentServiceFee: prometheus.NewGauge(prometheus.GaugeOpts{Name: "vastai_machine_earnings_current_service_fee"}),
		totalGPU:         prometheus.NewGauge(prometheus.GaugeOpts{Name: "vastai_machine_earnings_total_gpu"}),
	}
}

// Describe sends the metrics descriptions to the channel.
func (c *machineEarningsCollector) Describe(ch chan<- *prometheus.Desc) {
	c.currentBalance.Describe(ch)
	c.currentServiceFee.Describe(ch)
	c.totalGPU.Describe(ch)
}

// Collect sends the metrics values to the channel.
func (c *machineEarningsCollector) Collect(ch chan<- prometheus.Metric) {
	c.Update()

	c.currentBalance.Collect(ch)
	c.currentServiceFee.Collect(ch)
	c.totalGPU.Collect(ch)
}

// Update fetches data from the /machine-earnings endpoint and updates the metrics.
func (c *machineEarningsCollector) Update() {
	url := fmt.Sprintf("https://console.vast.ai/api/v0/users/me/machine-earnings?api_key=%s", c.apiKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Errorln("Failed to create new request:", err)
		return
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorln("Failed to fetch data from /machine-earnings:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorln("Failed to read response body:", err)
		return
	}

	log.Infof("Response body: %s", body)  // Added logging statement

	var data struct {
		Summary struct {
			TotalGPU float64 `json:"total_gpu"`
		} `json:"summary"`
		Current struct {
			Balance    float64 `json:"balance"`
			ServiceFee float64 `json:"service_fee"`
		} `json:"current"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		log.Errorln("Failed to parse JSON response:", err)
		return
	}

	c.currentBalance.Set(data.Current.Balance)
	c.currentServiceFee.Set(data.Current.ServiceFee)
	c.totalGPU.Set(data.Summary.TotalGPU)
}
