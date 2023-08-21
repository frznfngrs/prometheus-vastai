package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
)

type VastCollector struct {
	apiKey      string
	metrics     map[string]*prometheus.Desc
	totalMachines *prometheus.Desc
}

func NewVastCollector(apiKey string) *VastCollector {
	return &VastCollector{
		apiKey: apiKey,
		metrics: map[string]*prometheus.Desc{
			"earnings_current_balance": prometheus.NewDesc(
				"vastai_earnings_current_balance",
				"Current balance of machine earnings",
				nil, nil,
			),
			"earnings_total_gpu": prometheus.NewDesc(
				"vastai_earnings_total_gpu",
				"Total GPU earnings",
				nil, nil,
			),
			"machines_total": prometheus.NewDesc(
				"vastai_machines_total",
				"Total number of machines",
				nil, nil,
			),
			// Add more metrics here as needed
		},
		totalMachines: prometheus.NewDesc(
			"vastai_machines_total",
			"Total number of machines",
			nil, nil,
		),
	}
}

func (c *VastCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric
	}
	ch <- c.totalMachines
}

func (c *VastCollector) Collect(ch chan<- prometheus.Metric) {
	// Fetch data from machineEarnings endpoint
	earningsURL := fmt.Sprintf("https://console.vast.ai/api/v0/users/me/machine-earnings?api_key=%s", c.apiKey)
	resp, err := http.Get(earningsURL)
	if err != nil {
		log.Printf("Error fetching machine earnings data: %s", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var earningsData map[string]interface{}
	err = json.Unmarshal(bodyBytes, &earningsData)
	if err != nil {
		log.Printf("Error decoding machine earnings data: %s", err)
		return
	}
	log.Println("Successfully fetched machine earnings data.")
	currentBalance := earningsData["current"].(map[string]interface{})["balance"].(float64)
	totalGPU := earningsData["summary"].(map[string]interface{})["total_gpu"].(float64)

	ch <- prometheus.MustNewConstMetric(c.metrics["earnings_current_balance"], prometheus.GaugeValue, currentBalance)
	ch <- prometheus.MustNewConstMetric(c.metrics["earnings_total_gpu"], prometheus.GaugeValue, totalGPU)

	// Fetch data from Machines endpoint
	machinesURL := fmt.Sprintf("https://console.vast.ai/api/v0/machines?api_key=%s", c.apiKey)
	resp, err = http.Get(machinesURL)
	if err != nil {
		log.Printf("Error fetching machines data: %s", err)
		return
	}
	defer resp.Body.Close()

	var machinesData map[string][]interface{}
	err = json.Unmarshal(bodyBytes, &machinesData)
	if err != nil {
		log.Printf("Error decoding machines data: %s", err)
		return
	}
	log.Println("Successfully fetched machines data.")

	totalMachines := float64(len(machinesData["machines"]))
	ch <- prometheus.MustNewConstMetric(c.totalMachines, prometheus.GaugeValue, totalMachines)
}
