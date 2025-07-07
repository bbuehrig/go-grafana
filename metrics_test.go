package main

import (
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

type testService struct {
	Service
}

func (s *testService) initMetricsWithRegistry(reg prometheus.Registerer) {
	s.metrics.siteStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "site_status_test",
		Help: "Test: The summary of monitored sites and their response-codes",
	}, []string{"url"})
	reg.MustRegister(s.metrics.siteStatus)

	s.metrics.sites = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "sites_test",
		Help: "Test: The number of monitored sites",
	})
	reg.MustRegister(s.metrics.sites)

	s.metrics.offlineSites = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "offline_sites_test",
		Help: "Test: The number of offline sites",
	})
	reg.MustRegister(s.metrics.offlineSites)

	s.metrics.errorCounter = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "error_sites_test",
		Help: "Test: How long the sites are offline",
	}, []string{"url"})
	reg.MustRegister(s.metrics.errorCounter)
}

func TestInitMetricsFields(t *testing.T) {
	ts := &testService{}
	reg := prometheus.NewRegistry()
	ts.initMetricsWithRegistry(reg)

	if ts.metrics.siteStatus == nil || reflect.TypeOf(ts.metrics.siteStatus).String() != "*prometheus.GaugeVec" {
		t.Errorf("siteStatus not initialized or wrong type")
	}
	if ts.metrics.sites == nil {
		t.Errorf("sites not initialized")
	} else {
		if _, ok := ts.metrics.sites.(prometheus.Counter); !ok {
			t.Errorf("sites does not implement prometheus.Counter")
		}
	}
	if ts.metrics.offlineSites == nil {
		t.Errorf("offlineSites not initialized")
	} else {
		if _, ok := ts.metrics.offlineSites.(prometheus.Gauge); !ok {
			t.Errorf("offlineSites does not implement prometheus.Gauge")
		}
	}
	if ts.metrics.errorCounter == nil || reflect.TypeOf(ts.metrics.errorCounter).String() != "*prometheus.GaugeVec" {
		t.Errorf("errorCounter not initialized or wrong type")
	}
}
