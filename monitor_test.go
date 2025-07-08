package main

import (
	"errors"
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

type fakeMetrics struct {
	siteStatusSet map[string]float64
	offlineSites  int
	errorCounter  map[string]int
}

func newFakeMetrics() *fakeMetrics {
	return &fakeMetrics{
		siteStatusSet: make(map[string]float64),
		errorCounter:  make(map[string]int),
	}
}

func (f *fakeMetrics) siteStatusWith(url string, val float64) {
	f.siteStatusSet[url] = val
}
func (f *fakeMetrics) offlineSitesInc()                          { f.offlineSites++ }
func (f *fakeMetrics) errorCounterWithInc(url string)            { f.errorCounter[url]++ }
func (f *fakeMetrics) errorCounterWithSet(url string, v float64) { f.errorCounter[url] = int(v) }

// Mocks for Service dependencies

type mockEmailSender struct {
	lastSubject string
	lastBody    string
	calls       int
}

func (m *mockEmailSender) Send(subject, body string) error {
	m.lastSubject = subject
	m.lastBody = body
	m.calls++
	return nil
}

// Helper to create a Service with mocks
func newTestService() *Service {
	s := &Service{
		offlineMap:   make(map[string]bool),
		failureCount: make(map[string]int),
		emailSender:  &mockEmailSender{},
	}
	s.metrics.siteStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "test_site_status", Help: ""}, []string{"url"})
	s.metrics.offlineSites = prometheus.NewGauge(prometheus.GaugeOpts{Name: "test_offline_sites", Help: ""})
	s.metrics.errorCounter = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "test_error_counter", Help: ""}, []string{"url"})
	return s
}

func TestCheckSiteStatus_Error(t *testing.T) {
	s := newTestService()
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("fail")
	})}
	s.checkSiteStatus("https://fail.com", client)
	if !s.offlineMap["https://fail.com"] {
		t.Errorf("offlineMap not set for error")
	}
	me := s.emailSender.(*mockEmailSender)
	if me.calls != 1 {
		t.Errorf("expected 1 alert, got %d", me.calls)
	}
}

func TestCheckSiteStatus_Recovery(t *testing.T) {
	s := newTestService()
	s.offlineMap["https://ok.com"] = true
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	})}
	s.checkSiteStatus("https://ok.com", client)
	if s.offlineMap["https://ok.com"] {
		t.Errorf("offlineMap not reset on recovery")
	}
	me := s.emailSender.(*mockEmailSender)
	if me.calls != 1 {
		t.Errorf("expected 1 recovery alert, got %d", me.calls)
	}
}

func TestCheckSiteStatus_Threshold(t *testing.T) {
	s := newTestService()
	s.config.alertThreshold = 3
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("fail")
	})}
	for i := 0; i < 2; i++ {
		s.checkSiteStatus("https://fail.com", client)
	}
	me := s.emailSender.(*mockEmailSender)
	if me.calls != 0 {
		t.Errorf("expected no alert before threshold, got %d", me.calls)
	}
	s.checkSiteStatus("https://fail.com", client)
	if me.calls != 1 {
		t.Errorf("expected alert at threshold, got %d", me.calls)
	}
}

func TestCheckSiteStatus_ThresholdResetOnRecovery(t *testing.T) {
	s := newTestService()
	s.config.alertThreshold = 2
	failClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("fail")
	})}
	okClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
	})}
	s.checkSiteStatus("https://fail.com", failClient)
	s.checkSiteStatus("https://fail.com", failClient)
	me := s.emailSender.(*mockEmailSender)
	if me.calls != 1 {
		t.Errorf("expected alert at threshold, got %d", me.calls)
	}
	// Now recover
	s.checkSiteStatus("https://fail.com", okClient)
	if s.failureCount["https://fail.com"] != 0 {
		t.Errorf("expected failureCount reset on recovery, got %d", s.failureCount["https://fail.com"])
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
