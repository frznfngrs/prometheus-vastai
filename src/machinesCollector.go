package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type machinesCollector struct {
	apiKey           string
	currentBalance   prometheus.Gauge
	currentServiceFee prometheus.Gauge
	totalGPU          prometheus.Gauge
}

func NewMachinesCollector(apiKey string) *machinesCollector {
	return &machinesCollector{
		apiKey:           apiKey,
		currentBalance:   prometheus.NewGauge(prometheus.GaugeOpts{Name: "vastai_machine_earnings_current_balance"}),
		currentServiceFee: prometheus.NewGauge(prometheus.GaugeOpts{Name: "vastai_machine_earnings_current_service_fee"}),
		totalGPU:          prometheus.NewGauge(prometheus.GaugeOpts{Name: "vastai_machine_earnings_total_gpu"}),
	}
}

func (c *machinesCollector) Describe(ch chan<- *prometheus.Desc) {
	c.currentBalance.Describe(ch)
	c.currentServiceFee.Describe(ch)
	c.totalGPU.Describe(ch)
}

func (c *machinesCollector) Collect(ch chan<- prometheus.Metric) {
	c.Update()

	c.currentBalance.Collect(ch)
	c.currentServiceFee.Collect(ch)
	c.totalGPU.Collect(ch)
}

func (c *machinesCollector) Update() {
	req, err := http.NewRequest("GET", "https://console.vast.ai/api/v0/users/me/machine-earnings", nil)
	if err != nil {
		log.Errorln("Failed to create new request:", err)
		return
	}
	req.Header.Add("Authorization", "Bearer "+c.apiKey)

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

	var data struct {
		Current struct {
			Balance    float64 `json:"balance"`
			ServiceFee float64 `json:"service_fee"`
		} `json:"current"`
		Summary struct {
			TotalGPU float64 `json:"total_gpu"`
		} `json:"summary"`
	}

	if err := json.Unmarshal(body, &data); err != nil {
		log.Errorln("Failed to parse JSON response:", err)
		return
	}

	c.currentBalance.Set(data.Current.Balance)
	c.currentServiceFee.Set(data.Current.ServiceFee)
	c.totalGPU.Set(data.Summary.TotalGPU)
}
