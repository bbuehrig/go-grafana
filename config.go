package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const defaultCheckDurationTime = 51

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

	log.Println("Loaded configuration:")
	log.Printf("  URLs: %v", s.config.urls)
	log.Printf("  Check interval: %v", s.config.checkInterval)
	log.Printf("  SMTP server: %s:%s", s.config.smtpServer, s.config.smtpPort)
	log.Printf("  SMTP user: %s", s.config.smtpUser)
	log.Printf("  SMTP to: %s", s.config.smtpTo)
	log.Printf("  SMTP from: %s", s.config.smtpFrom)
}
