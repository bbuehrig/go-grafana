package main

import (
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const checkDurationTime = 51

type Service struct {
	metrics appMetrics
	config  appConfig
}

type appMetrics struct {
	siteStatus   *prometheus.GaugeVec
	sites        prometheus.Counter
	offlineSites prometheus.Gauge
	errorCounter *prometheus.GaugeVec
}

type appConfig struct {
	urls []string
}

func (s *Service) readConfig() {
	err := godotenv.Load("config/.env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	s.config.urls = strings.Split(os.Getenv("URLS"), ",")
	s.metrics.sites.Add(float64(len(s.config.urls)))
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

func (s *Service) recordMetrics() {
	go func() {
		for {
			s.metrics.offlineSites.Set(0)

			for _, url := range s.config.urls {
				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					s.metrics.siteStatus.With(prometheus.Labels{"url": url}).Set(0)
					s.metrics.offlineSites.Inc()
					s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Inc()
					continue
				}

				res, err := http.DefaultClient.Do(req)
				if err != nil {
					s.metrics.siteStatus.With(prometheus.Labels{"url": url}).Set(0)
					s.metrics.offlineSites.Inc()
					s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Inc()
					continue
				}

				defer res.Body.Close()

				s.metrics.siteStatus.Delete(prometheus.Labels{"url": url})

				if res == nil {
					s.metrics.siteStatus.With(prometheus.Labels{"url": url}).Set(0)
					s.metrics.offlineSites.Inc()
					s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Inc()
					continue
				}

				s.metrics.siteStatus.With(prometheus.Labels{"url": url}).Set(float64(res.StatusCode))

				if res.StatusCode < 200 || res.StatusCode >= 300 {
					s.metrics.offlineSites.Inc()
					s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Inc()
				} else {
					// no more errors - site es online again!
					s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Set(0)
				}
			}
			time.Sleep(checkDurationTime * time.Second)
		}
	}()
}

func newService() *Service {
	service := &Service{}
	service.initMetrics()
	service.readConfig()
	return service
}

func main() {
	service := newService()
	service.recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
