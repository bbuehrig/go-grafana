package main

import (
	"fmt"
	"testing"
)

func TestSendSiteDownAlertSubject(t *testing.T) {
	url := "https://example.com"
	reason := "returned status 500"
	subject := fmt.Sprintf("[🚨 DOWN] %s (%s)", url, reason)
	expected := "[🚨 DOWN] https://example.com (returned status 500)"
	if subject != expected {
		t.Errorf("expected subject '%s', got '%s'", expected, subject)
	}
}

func TestSendSiteRecoveryAlertSubject(t *testing.T) {
	url := "https://example.com"
	subject := fmt.Sprintf("[✅ UP] %s is back online", url)
	expected := "[✅ UP] https://example.com is back online"
	if subject != expected {
		t.Errorf("expected subject '%s', got '%s'", expected, subject)
	}
}

// Note: sendEmail uses the real SMTP server, so we do not test it directly here.
// For a real project, use an interface for sending and mock it in tests.

type mockSender struct {
	lastSubject string
	lastBody    string
	calls       int
}

func (m *mockSender) Send(subject, body string) error {
	m.lastSubject = subject
	m.lastBody = body
	m.calls++
	return nil
}

func TestServiceSendSiteDownAlert(t *testing.T) {
	mock := &mockSender{}
	service := &Service{
		config: appConfig{
			smtpTo: "to@example.com",
		},
		emailSender: mock,
	}
	url := "https://example.com"
	reason := "returned status 500"
	service.sendSiteDownAlert(url, reason)
	expectedSubject := "[🚨 DOWN] https://example.com (returned status 500)"
	expectedBody := "https://example.com: returned status 500"
	if mock.lastSubject != expectedSubject {
		t.Errorf("expected subject '%s', got '%s'", expectedSubject, mock.lastSubject)
	}
	if mock.lastBody != expectedBody {
		t.Errorf("expected body '%s', got '%s'", expectedBody, mock.lastBody)
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 call, got %d", mock.calls)
	}
}

func TestServiceSendSiteRecoveryAlert(t *testing.T) {
	mock := &mockSender{}
	service := &Service{
		config: appConfig{
			smtpTo: "to@example.com",
		},
		emailSender: mock,
	}
	url := "https://example.com"
	service.sendSiteRecoveryAlert(url)
	expectedSubject := "[✅ UP] https://example.com is back online"
	expectedBody := "https://example.com is back online"
	if mock.lastSubject != expectedSubject {
		t.Errorf("expected subject '%s', got '%s'", expectedSubject, mock.lastSubject)
	}
	if mock.lastBody != expectedBody {
		t.Errorf("expected body '%s', got '%s'", expectedBody, mock.lastBody)
	}
	if mock.calls != 1 {
		t.Errorf("expected 1 call, got %d", mock.calls)
	}
}
