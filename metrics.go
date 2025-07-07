package main

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

type appMetrics struct {
	siteStatus   *prometheus.GaugeVec
	sites        prometheus.Counter
	offlineSites prometheus.Gauge
	errorCounter *prometheus.GaugeVec
}

func (s *Service) initMetrics() {
	s.metrics.siteStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "site_status",
		Help: "The summary of monitored sites and their response-codes",
	}, []string{"url"})
	if err := prometheus.Register(s.metrics.siteStatus); err != nil && err.Error() != "duplicate metrics collector for site_status registration attempted" {
		log.Fatal(err)
	}

	s.metrics.sites = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sites",
		Help: "The number of monitored sites",
	})
	if err := prometheus.Register(s.metrics.sites); err != nil && err.Error() != "duplicate metrics collector for sites registration attempted" {
		log.Fatal(err)
	}

	s.metrics.offlineSites = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "offline_sites",
		Help: "The number of offline sites",
	})
	if err := prometheus.Register(s.metrics.offlineSites); err != nil && err.Error() != "duplicate metrics collector for offline sites registration attempted" {
		log.Fatal(err)
	}

	s.metrics.errorCounter = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "error_sites",
		Help: "How long the sites are offline",
	}, []string{"url"})
	if err := prometheus.Register(s.metrics.errorCounter); err != nil && err.Error() != "duplicate metrics collector for error counter for offline sites registration attempted" {
		log.Fatal(err)
	}
}
