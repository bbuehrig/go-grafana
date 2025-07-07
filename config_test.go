package main

import (
	"os"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func setupEnv(env map[string]string) func() {
	old := make(map[string]string)
	for k, v := range env {
		old[k] = os.Getenv(k)
		os.Setenv(k, v)
	}
	return func() {
		for k, v := range old {
			os.Setenv(k, v)
		}
	}
}

func TestReadConfigLoadsFields(t *testing.T) {
	cleanup := setupEnv(map[string]string{
		"URLS":           "https://a.com,https://b.com",
		"CHECK_INTERVAL": "42s",
		"SMTP_SERVER":    "smtp.example.com",
		"SMTP_PORT":      "587",
		"SMTP_USER":      "user",
		"SMTP_PASS":      "pass",
		"SMTP_TO":        "to@example.com",
		"SMTP_FROM":      "from@example.com",
	})
	defer cleanup()

	s := &Service{}
	s.metrics.sites = prometheus.NewCounter(prometheus.CounterOpts{Name: "test_sites"})
	s.readConfig()

	if len(s.config.urls) != 2 || s.config.urls[0] != "https://a.com" || s.config.urls[1] != "https://b.com" {
		t.Errorf("URLs not loaded correctly: %v", s.config.urls)
	}
	if s.config.checkInterval != 42*time.Second {
		t.Errorf("checkInterval not parsed: got %v", s.config.checkInterval)
	}
	if s.config.smtpServer != "smtp.example.com" {
		t.Errorf("smtpServer not loaded")
	}
	if s.config.smtpPort != "587" {
		t.Errorf("smtpPort not loaded")
	}
	if s.config.smtpUser != "user" {
		t.Errorf("smtpUser not loaded")
	}
	if s.config.smtpPass != "pass" {
		t.Errorf("smtpPass not loaded")
	}
	if s.config.smtpTo != "to@example.com" {
		t.Errorf("smtpTo not loaded")
	}
	if s.config.smtpFrom != "from@example.com" {
		t.Errorf("smtpFrom not loaded")
	}
}

func TestReadConfigDefaultInterval(t *testing.T) {
	cleanup := setupEnv(map[string]string{
		"URLS":           "https://a.com",
		"CHECK_INTERVAL": "",
	})
	defer cleanup()

	s := &Service{}
	s.metrics.sites = prometheus.NewCounter(prometheus.CounterOpts{Name: "test_sites_default"})
	s.readConfig()

	expected := defaultCheckDurationTime * time.Second
	if s.config.checkInterval != expected {
		t.Errorf("expected default interval %v, got %v", expected, s.config.checkInterval)
	}
}
