package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	apiKey := flag.String("api-key", "", "Vast.ai API key")
	listenAddress := flag.String("listen-address", ":8622", "Address to listen on for HTTP requests.")
	flag.Parse()

	if *apiKey == "" {
		fmt.Println("API key must be provided")
		os.Exit(1)
	}

	collector := NewVastCollector(*apiKey)
	prometheus.DefaultRegisterer.Unregister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	prometheus.DefaultRegisterer.Unregister(prometheus.NewGoCollector())
	prometheus.MustRegister(collector)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("<h1>Vast.ai Exporter</h1><p><a href='/metrics'>Metrics</a></p>"))
	})
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting vast.ai exporter on %s", *listenAddress)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
