package main

import (
	"testing"
)

func TestNewServiceInitializesFields(t *testing.T) {
	s := newService()
	if s.offlineMap == nil {
		t.Error("offlineMap not initialized")
	}
	if s.metrics.siteStatus == nil {
		t.Error("metrics.siteStatus not initialized")
	}
	if s.metrics.sites == nil {
		t.Error("metrics.sites not initialized")
	}
	if s.metrics.offlineSites == nil {
		t.Error("metrics.offlineSites not initialized")
	}
	if s.metrics.errorCounter == nil {
		t.Error("metrics.errorCounter not initialized")
	}
	if s.emailSender == nil {
		t.Error("emailSender not initialized")
	}
}
