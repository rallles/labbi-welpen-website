package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"labbi-app/internal/config"
	"labbi-app/internal/models"
)

func TestBuildContactMailMessageSanitizesHeaders(t *testing.T) {
	cfg := config.Config{
		SMTPUser:      "sender@example.invalid",
		ContactMailTo: "contact@example.invalid",
	}
	contact := models.Contact{
		Name:    "Alice\r\nBcc: attacker@example.invalid",
		Email:   "alice@example.invalid",
		Phone:   "+49 123",
		Message: "Hallo,\nbitte melden.",
	}

	message, err := buildContactMailMessage(cfg, contact)
	if err != nil {
		t.Fatalf("buildContactMailMessage() error = %v", err)
	}

	parts := strings.SplitN(string(message), "\r\n\r\n", 2)
	if len(parts) != 2 {
		t.Fatalf("message does not contain header/body separator: %q", string(message))
	}
	headers := parts[0]
	body := parts[1]

	requiredHeaders := []string{
		"From: ",
		"To: ",
		"Reply-To: ",
		"Subject: ",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}
	for _, header := range requiredHeaders {
		if !strings.Contains(headers, header) {
			t.Fatalf("headers %q do not contain %q", headers, header)
		}
	}
	if strings.Contains(headers, "\r\nBcc:") || strings.Contains(headers, "\nBcc:") {
		t.Fatalf("headers contain injected Bcc header: %q", headers)
	}
	if !strings.Contains(body, "Nachricht:") {
		t.Fatalf("body missing contact message section: %q", body)
	}
}

func TestBuildContactMailMessageRejectsInvalidReplyTo(t *testing.T) {
	cfg := config.Config{
		SMTPUser:      "sender@example.invalid",
		ContactMailTo: "contact@example.invalid",
	}
	contact := models.Contact{
		Name:    "Alice",
		Email:   "not an email",
		Message: "Hallo zusammen",
	}

	if _, err := buildContactMailMessage(cfg, contact); err == nil {
		t.Fatal("buildContactMailMessage() error = nil, want error")
	}
}

func TestClientIPPrefersValidXRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/contact", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	req.Header.Set("X-Real-IP", "198.51.100.7")
	req.Header.Set("X-Forwarded-For", "192.0.2.1")

	if got := clientIP(req); got != "198.51.100.7" {
		t.Fatalf("clientIP() = %q, want X-Real-IP", got)
	}
}

func TestClientIPIgnoresSpoofedForwardedChain(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/contact", nil)
	req.RemoteAddr = "203.0.113.10:12345"
	req.Header.Set("X-Forwarded-For", "198.51.100.7, 192.0.2.1")

	if got := clientIP(req); got != "203.0.113.10" {
		t.Fatalf("clientIP() = %q, want RemoteAddr fallback", got)
	}
}

func TestClientIPUsesSingleValidForwardedForWhenNeeded(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/contact", nil)
	req.RemoteAddr = "bad-remote-addr"
	req.Header.Set("X-Forwarded-For", "198.51.100.7")

	if got := clientIP(req); got != "198.51.100.7" {
		t.Fatalf("clientIP() = %q, want single valid X-Forwarded-For", got)
	}
}

func TestClientIPFallsBackToUnknown(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/contact", nil)
	req.RemoteAddr = "bad-remote-addr"
	req.Header.Set("X-Real-IP", "not-an-ip")
	req.Header.Set("X-Forwarded-For", "also-not-an-ip")

	if got := clientIP(req); got != "unknown" {
		t.Fatalf("clientIP() = %q, want unknown", got)
	}
}
