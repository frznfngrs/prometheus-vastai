package main

import (
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	listenAddress = kingpin.Flag(
		"listen",
		"Address to listen on.",
	).Default(":8622").String()
	apiKey = kingpin.Flag(
		"key",
		"Vast.ai API key",
	).Default("").String()
	updateInterval = kingpin.Flag(
		"update-interval",
		"How often to query Vast.ai for updates",
	).Default("1m").Duration()
	stateDir = kingpin.Flag(
		"state-dir",
		"Path to store state files (default $HOME)",
	).String()
)

func main() {
	kingpin.Version(version.Print("vastai_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting vast.ai exporter")

	if *stateDir == "" {
		*stateDir = os.Getenv("HOME")
	}
	if *stateDir == "" {
		*stateDir = "/tmp"
	}

	log.Infoln("Reading initial Vast.ai info (may take a minute)")

	machinesCollector := NewMachinesCollector(*apiKey)
	machineEarningsCollector := NewMachineEarningsCollector(*apiKey)

	registry := prometheus.NewRegistry()
	registry.MustRegister(machinesCollector)
	registry.MustRegister(machineEarningsCollector)

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Write([]byte(`<html><head><title>Vast.ai Exporter</title></head><body><h1>Vast.ai Exporter</h1>`))
		w.Write([]byte(`<a href="metrics">Metrics</a>`))
		w.Write([]byte(`</body></html>`))
	})

	go func() {
		for {
			time.Sleep(*updateInterval)
			machinesCollector.Update()          // updated method name
			machineEarningsCollector.Update()
		}
	}()

	log.Infoln("Listening on", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
