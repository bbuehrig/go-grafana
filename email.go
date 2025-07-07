package main

import (
	"fmt"
	"log"
	"net/smtp"
)

func sendEmail(cfg appConfig, subject, body string) error {
	auth := smtp.PlainAuth("", cfg.smtpUser, cfg.smtpPass, cfg.smtpServer)
	addr := fmt.Sprintf("%s:%s", cfg.smtpServer, cfg.smtpPort)
	msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", cfg.smtpTo, subject, body))
	return smtp.SendMail(addr, auth, cfg.smtpFrom, []string{cfg.smtpTo}, msg)
}

func (s *Service) sendSiteDownAlert(url, reason string) {
	log.Printf("Sending email: subject='%s' to='%s' (reason: %s)", "Site Down", s.config.smtpTo, reason)
	if err := sendEmail(s.config, "Site Down", fmt.Sprintf("%s: %s", url, reason)); err != nil {
		log.Printf("Failed to send email: subject='%s' to='%s': %v", "Site Down", s.config.smtpTo, err)
	} else {
		log.Printf("Alert sent: %s - %s", url, reason)
	}
}

func (s *Service) sendSiteRecoveryAlert(url string) {
	log.Printf("Sending email: subject='%s' to='%s' (reason: recovery)", "Site Recovered", s.config.smtpTo)
	if err := sendEmail(s.config, "Site Recovered", fmt.Sprintf("%s is back online", url)); err != nil {
		log.Printf("Failed to send email: subject='%s' to='%s': %v", "Site Recovered", s.config.smtpTo, err)
	} else {
		log.Printf("Recovery alert sent: %s is back online", url)
	}
}
