package main

import (
	"fmt"
	"log"

	"net/smtp"

	"github.com/jordan-wright/email"
)

func sendEmail(cfg appConfig, subject, body string) error {
	e := email.NewEmail()
	e.From = cfg.smtpFrom
	e.To = []string{cfg.smtpTo}
	e.Subject = subject
	e.Text = []byte(body)

	auth := smtp.PlainAuth("", cfg.smtpUser, cfg.smtpPass, cfg.smtpServer)
	addr := fmt.Sprintf("%s:%s", cfg.smtpServer, cfg.smtpPort)
	return e.Send(addr, auth)
}

func (s *Service) sendSiteDownAlert(url, reason string) {
	subject := fmt.Sprintf("[🚨 DOWN] %s (%s)", url, reason)
	log.Printf("Sending email: subject='%s' to='%s' (reason: %s)", subject, s.config.smtpTo, reason)
	if err := sendEmail(s.config, subject, fmt.Sprintf("%s: %s", url, reason)); err != nil {
		log.Printf("Failed to send email: subject='%s' to='%s': %v", subject, s.config.smtpTo, err)
	} else {
		log.Printf("Alert sent: %s - %s", url, reason)
	}
}

func (s *Service) sendSiteRecoveryAlert(url string) {
	subject := fmt.Sprintf("[✅ UP] %s is back online", url)
	log.Printf("Sending email: subject='%s' to='%s' (reason: recovery)", subject, s.config.smtpTo)
	if err := sendEmail(s.config, subject, fmt.Sprintf("%s is back online", url)); err != nil {
		log.Printf("Failed to send email: subject='%s' to='%s': %v", subject, s.config.smtpTo, err)
	} else {
		log.Printf("Recovery alert sent: %s is back online", url)
	}
}
