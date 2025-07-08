package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func (s *Service) recordMetrics(ctx context.Context) {
	go func() {
		client := &http.Client{Timeout: 10 * time.Second}

		for {
			start := time.Now()
			select {
			case <-ctx.Done():
				log.Println("Shutting down monitoring goroutine...")
				return
			default:
			}

			s.metrics.offlineSites.Set(0)

			var wg sync.WaitGroup
			for _, url := range s.config.urls {
				wg.Add(1)
				go func(url string) {
					defer wg.Done()
					s.checkSiteStatus(url, client)
				}(url)
			}
			wg.Wait()

			elapsed := time.Since(start)
			if elapsed < s.config.checkInterval {
				time.Sleep(s.config.checkInterval - elapsed)
			}
		}
	}()
}

func (s *Service) checkSiteStatus(url string, client *http.Client) {
	// Create HTTP request
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		s.handleSiteError(url, fmt.Sprintf("unreachable: %v", err))
		return
	}

	// Execute request
	res, err := client.Do(req)
	if err != nil {
		s.handleSiteError(url, fmt.Sprintf("unreachable: %v", err))
		return
	}
	defer res.Body.Close()

	// Check for nil response
	if res == nil {
		s.handleSiteError(url, "returned nil response")
		return
	}

	// Update metrics
	s.metrics.siteStatus.Delete(prometheus.Labels{"url": url})
	s.metrics.siteStatus.With(prometheus.Labels{"url": url}).Set(float64(res.StatusCode))

	// Check status code
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		s.handleSiteError(url, fmt.Sprintf("returned status %d", res.StatusCode))
	} else {
		// Site is healthy
		s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Set(0)
		s.handleSiteRecovery(url)
	}
}

func (s *Service) handleSiteError(url, reason string) {
	s.metrics.siteStatus.With(prometheus.Labels{"url": url}).Set(0)
	s.metrics.offlineSites.Inc()
	s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Inc()

	s.mu.Lock()
	alreadyOffline := s.offlineMap[url]
	s.failureCount[url]++
	shouldAlert := !alreadyOffline && s.failureCount[url] >= s.config.alertThreshold
	if shouldAlert {
		s.offlineMap[url] = true
	}
	s.mu.Unlock()

	if shouldAlert {
		s.sendSiteDownAlert(url, reason)
	}
}

func (s *Service) handleSiteRecovery(url string) {
	s.mu.Lock()
	wasOffline := s.offlineMap[url]
	if wasOffline {
		s.offlineMap[url] = false
	}
	s.failureCount[url] = 0 // Reset failure count on recovery
	s.mu.Unlock()

	if wasOffline {
		s.sendSiteRecoveryAlert(url)
	}
}
