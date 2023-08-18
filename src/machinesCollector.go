package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type machinesCollector struct {
	apiKey        string
	totalMachines prometheus.Gauge
	numGPUs       *prometheus.GaugeVec
	totalFlops    *prometheus.GaugeVec
	gpuRAM        *prometheus.GaugeVec
	cpuRAM        *prometheus.GaugeVec
	cpuCores      *prometheus.GaugeVec
}

func NewMachinesCollector(apiKey string) *machinesCollector {
	return &machinesCollector{
		apiKey: apiKey,
		totalMachines: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "vastai_machines_total",
			Help: "Total number of machines.",
		}),
		numGPUs: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "vastai_machine_num_gpus",
			Help: "Number of GPUs in a machine.",
		}, []string{"machine_id", "hostname"}),
		totalFlops: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "vastai_machine_total_flops",
			Help: "Total FLOPs of a machine.",
		}, []string{"machine_id", "hostname"}),
		gpuRAM: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "vastai_machine_gpu_ram",
			Help: "GPU RAM of a machine.",
		}, []string{"machine_id", "hostname"}),
		cpuRAM: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "vastai_machine_cpu_ram",
			Help: "CPU RAM of a machine.",
		}, []string{"machine_id", "hostname"}),
		cpuCores: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "vastai_machine_cpu_cores",
			Help: "Number of CPU cores in a machine.",
		}, []string{"machine_id", "hostname"}),
	}
}

func (c *machinesCollector) Describe(ch chan<- *prometheus.Desc) {
	c.totalMachines.Describe(ch)
	c.numGPUs.Describe(ch)
	c.totalFlops.Describe(ch)
	c.gpuRAM.Describe(ch)
	c.cpuRAM.Describe(ch)
	c.cpuCores.Describe(ch)
}

func (c *machinesCollector) Collect(ch chan<- prometheus.Metric) {
	c.Update()
	c.totalMachines.Collect(ch)
	c.numGPUs.Collect(ch)
	c.totalFlops.Collect(ch)
	c.gpuRAM.Collect(ch)
	c.cpuRAM.Collect(ch)
	c.cpuCores.Collect(ch)
}

func (c *machinesCollector) Update() {
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
