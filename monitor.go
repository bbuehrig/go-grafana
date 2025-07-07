package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func (s *Service) recordMetrics(ctx context.Context) {
	go func() {
		offlineMap := make(map[string]bool)
		client := &http.Client{Timeout: 10 * time.Second}

		for {
			select {
			case <-ctx.Done():
				log.Println("Shutting down monitoring goroutine...")
				return
			default:
			}

			s.metrics.offlineSites.Set(0)

			for _, url := range s.config.urls {
				s.checkSiteStatus(url, client, offlineMap)
			}

			time.Sleep(s.config.checkInterval)
		}
	}()
}

func (s *Service) checkSiteStatus(url string, client *http.Client, offlineMap map[string]bool) {
	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		s.handleSiteError(url, fmt.Sprintf("unreachable: %v", err), offlineMap)
		return
	}

	// Execute request
	res, err := client.Do(req)
	if err != nil {
		s.handleSiteError(url, fmt.Sprintf("unreachable: %v", err), offlineMap)
		return
	}
	defer res.Body.Close()

	// Check for nil response
	if res == nil {
		s.handleSiteError(url, "returned nil response", offlineMap)
		return
	}

	// Update metrics
	s.metrics.siteStatus.Delete(prometheus.Labels{"url": url})
	s.metrics.siteStatus.With(prometheus.Labels{"url": url}).Set(float64(res.StatusCode))

	// Check status code
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		s.handleSiteError(url, fmt.Sprintf("returned status %d", res.StatusCode), offlineMap)
	} else {
		// Site is healthy
		s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Set(0)
		if offlineMap[url] {
			s.sendSiteRecoveryAlert(url)
			offlineMap[url] = false
		}
	}
}

func (s *Service) handleSiteError(url, reason string, offlineMap map[string]bool) {
	s.metrics.siteStatus.With(prometheus.Labels{"url": url}).Set(0)
	s.metrics.offlineSites.Inc()
	s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Inc()

	if !offlineMap[url] {
		s.sendSiteDownAlert(url, reason)
		offlineMap[url] = true
	}
}
