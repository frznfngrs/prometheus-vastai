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
				"Total GPU earnings
