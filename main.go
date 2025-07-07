package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"net/smtp"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const defaultCheckDurationTime = 51

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
	urls          []string
	checkInterval time.Duration
	smtpServer    string
	smtpPort      string
	smtpUser      string
	smtpPass      string
	smtpTo        string
	smtpFrom      string
}

func (s *Service) readConfig() {
	err := godotenv.Load("config/.env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}
	s.config.urls = strings.Split(os.Getenv("URLS"), ",")
	interval := os.Getenv("CHECK_INTERVAL")
	if interval == "" {
		s.config.checkInterval = defaultCheckDurationTime * time.Second
	} else {
		dur, err := time.ParseDuration(interval)
		if err != nil {
			log.Fatalf("Invalid CHECK_INTERVAL: %s", err)
		}
		s.config.checkInterval = dur
	}
	s.config.smtpServer = os.Getenv("SMTP_SERVER")
	s.config.smtpPort = os.Getenv("SMTP_PORT")
	s.config.smtpUser = os.Getenv("SMTP_USER")
	s.config.smtpPass = os.Getenv("SMTP_PASS")
	s.config.smtpTo = os.Getenv("SMTP_TO")
	s.config.smtpFrom = os.Getenv("SMTP_FROM")
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

func sendEmail(cfg appConfig, subject, body string) error {
	auth := smtp.PlainAuth("", cfg.smtpUser, cfg.smtpPass, cfg.smtpServer)
	addr := fmt.Sprintf("%s:%s", cfg.smtpServer, cfg.smtpPort)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", cfg.smtpTo, subject, body))
	return smtp.SendMail(addr, auth, cfg.smtpFrom, []string{cfg.smtpTo}, msg)
}

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
				req, err := http.NewRequest(http.MethodGet, url, nil)
				if err != nil {
					s.metrics.siteStatus.With(prometheus.Labels{"url": url}).Set(0)
					s.metrics.offlineSites.Inc()
					s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Inc()
					if !offlineMap[url] {
						log.Printf("Sending email: subject='%s' to='%s' (reason: unreachable)", "Site Down", s.config.smtpTo)
						sendEmail(s.config, "Site Down", fmt.Sprintf("%s is unreachable: %v", url, err))
						log.Printf("Alert sent: %s is unreachable", url)
						offlineMap[url] = true
					}
					continue
				}
				res, err := client.Do(req)
				if err != nil {
					s.metrics.siteStatus.With(prometheus.Labels{"url": url}).Set(0)
					s.metrics.offlineSites.Inc()
					s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Inc()
					if !offlineMap[url] {
						log.Printf("Sending email: subject='%s' to='%s' (reason: unreachable)", "Site Down", s.config.smtpTo)
						sendEmail(s.config, "Site Down", fmt.Sprintf("%s is unreachable: %v", url, err))
						log.Printf("Alert sent: %s is unreachable", url)
						offlineMap[url] = true
					}
					continue
				}
				defer res.Body.Close()
				s.metrics.siteStatus.Delete(prometheus.Labels{"url": url})
				if res == nil {
					s.metrics.siteStatus.With(prometheus.Labels{"url": url}).Set(0)
					s.metrics.offlineSites.Inc()
					s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Inc()
					if !offlineMap[url] {
						log.Printf("Sending email: subject='%s' to='%s' (reason: nil response)", "Site Down", s.config.smtpTo)
						sendEmail(s.config, "Site Down", fmt.Sprintf("%s returned nil response", url))
						log.Printf("Alert sent: %s returned nil response", url)
						offlineMap[url] = true
					}
					continue
				}
				s.metrics.siteStatus.With(prometheus.Labels{"url": url}).Set(float64(res.StatusCode))
				if res.StatusCode < 200 || res.StatusCode >= 300 {
					s.metrics.offlineSites.Inc()
					s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Inc()
					if !offlineMap[url] {
						log.Printf("Sending email: subject='%s' to='%s' (reason: status %d)", "Site Down", s.config.smtpTo, res.StatusCode)
						sendEmail(s.config, "Site Down", fmt.Sprintf("%s returned status %d", url, res.StatusCode))
						log.Printf("Alert sent: %s returned status %d", url, res.StatusCode)
						offlineMap[url] = true
					}
				} else {
					s.metrics.errorCounter.With(prometheus.Labels{"url": url}).Set(0)
					if offlineMap[url] {
						log.Printf("Sending email: subject='%s' to='%s' (reason: recovery)", "Site Recovered", s.config.smtpTo)
						sendEmail(s.config, "Site Recovered", fmt.Sprintf("%s is back online", url))
						log.Printf("Recovery alert sent: %s is back online", url)
						offlineMap[url] = false
					}
				}
			}
			time.Sleep(s.config.checkInterval)
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	service.recordMetrics(ctx)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err := http.ListenAndServe(":2112", nil); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	<-ctx.Done()
	log.Println("Shutting down...")
}
