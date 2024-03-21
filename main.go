package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const checkDurationTime = 10

type Service struct {
	metrics appMetrics
	config  appConfig
}

type appMetrics struct {
	siteStatus *prometheus.GaugeVec
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
}

func (s *Service) initMetrics() {
	s.metrics.siteStatus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "site_status",
		Help: "The summary of monitored sites and their response-codes",
	}, []string{"url", "httpcode", "message"})
	err := prometheus.Register(s.metrics.siteStatus)
	if err != nil && err.Error() != "duplicate metrics collector registration attempted" {
		log.Fatal(err)
	}

}

func (s *Service) recordMetrics() {
	go func() {
		for {
			s.metrics.siteStatus.Reset()

			for _, url := range s.config.urls {
				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					s.metrics.siteStatus.With(prometheus.Labels{"url": url, "httpcode": "0", "message": err.Error()}).Inc()
					continue
				}

				res, err := http.DefaultClient.Do(req)
				if err != nil {
					s.metrics.siteStatus.With(prometheus.Labels{"url": url, "httpcode": "0", "message": err.Error()}).Inc()
					continue
				}

				if res == nil {
					s.metrics.siteStatus.With(prometheus.Labels{"url": url, "httpcode": "0", "message": "response is nil"}).Inc()
					continue
				}

				s.metrics.siteStatus.With(prometheus.Labels{
					"url":      url,
					"httpcode": fmt.Sprintf("%d", res.StatusCode),
					"message":  res.Status,
				}).Inc()
			}
			time.Sleep(checkDurationTime * time.Second)
		}
	}()
}

func newService() *Service {
	service := &Service{}
	service.initMetrics()
	return service
}

func main() {
	service := newService()
	service.readConfig()
	service.recordMetrics()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
