package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/mail"
	"strings"
	"testing"
	"time"

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

func TestBuildContactMailMessageNameCRLFDoesNotInjectHeader(t *testing.T) {
	cfg := config.Config{
		SMTPUser:      "sender@example.invalid",
		ContactMailTo: "contact@example.invalid",
	}
	contact := models.Contact{
		Name:    "Alice\r\nBcc: attacker@example.invalid\r\nX-Injected: yes",
		Email:   "alice@example.invalid",
		Phone:   "+49 123",
		Message: "Hallo zusammen",
	}

	message, err := buildContactMailMessage(cfg, contact)
	if err != nil {
		t.Fatalf("buildContactMailMessage() error = %v", err)
	}

	headers, _ := splitMailMessage(t, message)
	for _, line := range strings.Split(headers, "\r\n") {
		if strings.ContainsAny(line, "\r\n") {
			t.Fatalf("header line contains raw CR/LF: %q", line)
		}
		name, _, ok := strings.Cut(line, ":")
		if !ok {
			t.Fatalf("malformed header line: %q", line)
		}
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "bcc", "x-injected":
			t.Fatalf("headers contain injected %s header: %q", name, headers)
		}
	}
	if strings.Contains(headers, "\nBcc:") || strings.Contains(headers, "\nX-Injected:") {
		t.Fatalf("headers contain injected header line: %q", headers)
	}
}

func TestBuildContactMailMessageSanitizesSubject(t *testing.T) {
	cfg := config.Config{
		SMTPUser:      "sender@example.invalid",
		ContactMailTo: "contact@example.invalid",
	}
	contact := models.Contact{
		Name:    "Alice\r\nCc: attacker@example.invalid",
		Email:   "alice@example.invalid",
		Message: "Hallo zusammen",
	}

	message, err := buildContactMailMessage(cfg, contact)
	if err != nil {
		t.Fatalf("buildContactMailMessage() error = %v", err)
	}

	headerPart := strings.SplitN(string(message), "\r\n\r\n", 2)[0]
	var subjectLine string
	for _, line := range strings.Split(headerPart, "\r\n") {
		if strings.HasPrefix(line, "Subject: ") {
			subjectLine = line
			break
		}
	}
	if subjectLine == "" {
		t.Fatalf("Subject header not found in %q", headerPart)
	}
	if strings.Contains(subjectLine, "\n") || strings.Contains(subjectLine, "\r") {
		t.Fatalf("subject header contains CR/LF: %q", subjectLine)
	}
	if strings.Contains(subjectLine, "\r\nCc:") || strings.Contains(subjectLine, "\nCc:") {
		t.Fatalf("subject header contains injected Cc header: %q", subjectLine)
	}
	if !strings.Contains(subjectLine, "Neue Kontaktanfrage von Alice Cc attacker@example.invalid") {
		t.Fatalf("unexpected subject line: %q", subjectLine)
	}
}

func TestBuildContactMailMessageSetsReplyTo(t *testing.T) {
	cfg := config.Config{
		SMTPUser:      "sender@example.invalid",
		ContactMailTo: "contact@example.invalid",
	}
	contact := models.Contact{
		Name:    "Alice Example",
		Email:   "alice@example.invalid",
		Message: "Hallo zusammen",
	}

	message, err := buildContactMailMessage(cfg, contact)
	if err != nil {
		t.Fatalf("buildContactMailMessage() error = %v", err)
	}

	expectedReplyTo := "Reply-To: " + (&mail.Address{
		Name:    sanitizeHeaderValue(contact.Name),
		Address: contact.Email,
	}).String()

	headerPart := strings.SplitN(string(message), "\r\n\r\n", 2)[0]
	if !strings.Contains(headerPart, expectedReplyTo) {
		t.Fatalf("headers %q do not contain expected Reply-To %q", headerPart, expectedReplyTo)
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

func TestContactRateLimiterRemovesExpiredIPKeys(t *testing.T) {
	now := time.Now()
	limiter := &contactRateLimiter{
		requests: map[string][]time.Time{
			"198.51.100.7": []time.Time{now.Add(-2 * contactRateLimitWindow)},
			"203.0.113.10": []time.Time{now},
		},
		lastCleanup: now.Add(-2 * contactRateLimitCleanupInterval),
	}

	if !limiter.Allow("192.0.2.55") {
		t.Fatal("Allow() = false, want true")
	}
	if _, ok := limiter.requests["198.51.100.7"]; ok {
		t.Fatalf("expired IP key was not removed: %#v", limiter.requests)
	}
	if _, ok := limiter.requests["203.0.113.10"]; !ok {
		t.Fatalf("active IP key was removed: %#v", limiter.requests)
	}
}

func TestContactHandlerRejectsOversizedPostBody(t *testing.T) {
	body := "message=" + strings.Repeat("x", maxContactFormSize)
	request := httptest.NewRequest(http.MethodPost, "/contact", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()

	ContactHandler(response, request, config.Config{}, nil)

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body = %q", response.Code, http.StatusRequestEntityTooLarge, response.Body.String())
	}
}

func splitMailMessage(t *testing.T, message []byte) (string, string) {
	t.Helper()

	parts := strings.Split(string(message), "\r\n\r\n")
	if len(parts) != 2 {
		t.Fatalf("message must contain exactly one CRLF header/body separator, got %d parts: %q", len(parts), string(message))
	}
	return parts[0], parts[1]
}
