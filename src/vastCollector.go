package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
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

func (c *VastCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.metrics {
		ch <- metric
	}
}

func (c *VastCollector) Collect(ch chan<- prometheus.Metric) {
	c.fetchMachineEarnings(ch)
	// Call other fetch methods as you add them
}
