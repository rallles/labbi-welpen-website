package handlers

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/netip"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"labbi-app/internal/config"
	"labbi-app/internal/models"
	"labbi-app/internal/repository"
	"labbi-app/internal/validation"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const (
	maxContactFormSize              = 64 << 10
	contactRateLimitWindow          = 10 * time.Minute
	contactRateLimitCleanupInterval = 5 * time.Minute
	contactRateLimitMax             = 5
)

type ContactPageData struct {
	Form   validation.ContactForm
	Errors []string
}

type ContactResultData struct {
	Name           string
	Saved          bool
	MailConfigured bool
	MailSent       bool
	MailFailed     bool
}

type contactRateLimiter struct {
	mu          sync.Mutex
	requests    map[string][]time.Time
	lastCleanup time.Time
}

var defaultContactRateLimiter = &contactRateLimiter{requests: make(map[string][]time.Time)}

var errSMTPNotConfigured = errors.New("smtp_not_configured")

// ContactHandler zeigt das Kontaktformular (GET) und verarbeitet es (POST).
func ContactHandler(w http.ResponseWriter, r *http.Request, cfg config.Config, driver neo4j.DriverWithContext) {
	switch r.Method {
	case http.MethodGet:
		renderContactForm(w, validation.ContactForm{}, nil)
	case http.MethodPost:
		handleContactPost(w, r, cfg, driver)
	default:
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
	}
}

func handleContactPost(w http.ResponseWriter, r *http.Request, cfg config.Config, driver neo4j.DriverWithContext) {
	r.Body = http.MaxBytesReader(w, r.Body, maxContactFormSize)
	if err := r.ParseForm(); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			http.Error(w, "Kontaktformular ist zu groß", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Ungültige Eingaben", http.StatusBadRequest)
		return
	}

	form := validation.ContactFormFromValues(r.Form)
	if form.Website != "" {
		log.Printf("Kontakt-Honeypot ausgelöst von %s", clientIP(r))
		renderTemplate(w, "contact_result.html", ContactResultData{Saved: true, MailConfigured: smtpConfigured(cfg)})
		return
	}
	if !defaultContactRateLimiter.Allow(clientIP(r)) {
		log.Printf("Kontakt-Rate-Limit erreicht für %s", clientIP(r))
		renderTemplate(w, "contact_result.html", ContactResultData{Saved: true, MailConfigured: smtpConfigured(cfg)})
		return
	}

	if errs := validation.ValidateContactForm(form); len(errs) > 0 {
		renderContactForm(w, form, errs)
		return
	}

	contact := models.Contact{
		ID:        uuid.NewString(),
		Name:      form.Name,
		Email:     form.Email,
		Phone:     form.Phone,
		Message:   form.Message,
		CreatedAt: time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	contactRepo := repository.NewContactRepository(driver)
	if err := contactRepo.Create(ctx, contact); err != nil {
		log.Printf("Kontaktanfrage konnte nicht gespeichert werden: %v", err)
		renderContactForm(w, form, []string{"Ihre Anfrage konnte gerade nicht gespeichert werden. Bitte versuchen Sie es später erneut."})
		return
	}

	mailConfigured := smtpConfigured(cfg)
	mailErr := sendContactMail(cfg, contact)
	contact.MailSent = mailErr == nil && mailConfigured
	contact.MailError = sanitizeMailError(mailErr)
	if err := contactRepo.UpdateMailStatus(ctx, contact.ID, contact.MailSent, contact.MailError); err != nil {
		log.Printf("Mailstatus konnte nicht aktualisiert werden: %v", err)
	}
	if mailErr != nil && mailConfigured {
		log.Printf("E-Mail-Benachrichtigung fehlgeschlagen")
	}

	renderTemplate(w, "contact_result.html", ContactResultData{
		Name:           contact.Name,
		Saved:          true,
		MailConfigured: mailConfigured,
		MailSent:       contact.MailSent,
		MailFailed:     mailErr != nil && mailConfigured,
	})
}

func renderContactForm(w http.ResponseWriter, form validation.ContactForm, errors []string) {
	if err := renderTemplate(w, "contact.html", ContactPageData{Form: form, Errors: errors}); err != nil {
		log.Printf("Kontaktformular konnte nicht gerendert werden: %v", err)
	}
}

func (l *contactRateLimiter) Allow(ip string) bool {
	if ip == "" {
		ip = "unknown"
	}
	now := time.Now()
	cutoff := now.Add(-contactRateLimitWindow)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.requests == nil {
		l.requests = make(map[string][]time.Time)
	}
	if l.lastCleanup.IsZero() || now.Sub(l.lastCleanup) >= contactRateLimitCleanupInterval {
		l.cleanupLocked(cutoff)
		l.lastCleanup = now
	}

	items := l.requests[ip]
	kept := items[:0]
	for _, seenAt := range items {
		if seenAt.After(cutoff) {
			kept = append(kept, seenAt)
		}
	}
	if len(kept) >= contactRateLimitMax {
		l.requests[ip] = kept
		return false
	}
	l.requests[ip] = append(kept, now)
	return true
}

func (l *contactRateLimiter) cleanupLocked(cutoff time.Time) {
	for ip, items := range l.requests {
		kept := items[:0]
		for _, seenAt := range items {
			if seenAt.After(cutoff) {
				kept = append(kept, seenAt)
			}
		}
		if len(kept) == 0 {
			delete(l.requests, ip)
			continue
		}
		l.requests[ip] = kept
	}
}

func clientIP(r *http.Request) string {
	if ip, ok := validHeaderIP(r.Header.Get("X-Real-IP")); ok {
		return ip
	}
	if ip, ok := validHeaderIP(r.Header.Get("X-Forwarded-For")); ok {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	if ip, ok := validHeaderIP(host); ok {
		return ip
	}
	return "unknown"
}

func validHeaderIP(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if value == "" || strings.Contains(value, ",") {
		return "", false
	}
	addr, err := netip.ParseAddr(value)
	if err != nil {
		return "", false
	}
	return addr.String(), true
}

func smtpConfigured(cfg config.Config) bool {
	return strings.TrimSpace(cfg.SMTPHost) != "" && strings.TrimSpace(cfg.SMTPPort) != "" &&
		strings.TrimSpace(cfg.SMTPUser) != "" && strings.TrimSpace(cfg.SMTPPassword) != "" &&
		strings.TrimSpace(cfg.ContactMailTo) != ""
}

// sendContactMail versendet eine Benachrichtigungs-E-Mail.
func sendContactMail(cfg config.Config, contact models.Contact) error {
	if !smtpConfigured(cfg) {
		return errSMTPNotConfigured
	}

	fromAddress, err := parseHeaderAddress(cfg.SMTPUser, "invalid_from")
	if err != nil {
		return err
	}
	toAddress, err := parseHeaderAddress(cfg.ContactMailTo, "invalid_to")
	if err != nil {
		return err
	}

	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	msgData, err := buildContactMailMessage(cfg, contact)
	if err != nil {
		return err
	}

	return smtp.SendMail(cfg.SMTPHost+":"+cfg.SMTPPort, auth, fromAddress.Address, []string{toAddress.Address}, msgData)
}

func buildContactMailMessage(cfg config.Config, contact models.Contact) ([]byte, error) {
	fromAddress, err := parseHeaderAddress(cfg.SMTPUser, "invalid_from")
	if err != nil {
		return nil, err
	}
	toAddress, err := parseHeaderAddress(cfg.ContactMailTo, "invalid_to")
	if err != nil {
		return nil, err
	}
	replyTo, err := parseHeaderAddress(contact.Email, "invalid_reply_to")
	if err != nil {
		return nil, err
	}

	from := mail.Address{
		Name:    "Labbi-Welpen Kontaktformular",
		Address: fromAddress.Address,
	}
	to := mail.Address{Address: toAddress.Address}

	headers := []struct {
		key   string
		value string
	}{
		{"From", from.String()},
		{"To", to.String()},
		{"Reply-To", (&mail.Address{Name: sanitizeHeaderValue(contact.Name), Address: replyTo.Address}).String()},
		{"Subject", sanitizeHeaderValue("Neue Kontaktanfrage von " + contact.Name)},
		{"MIME-Version", "1.0"},
		{"Content-Type", "text/plain; charset=UTF-8"},
	}

	var msg bytes.Buffer
	for _, header := range headers {
		msg.WriteString(header.key)
		msg.WriteString(": ")
		msg.WriteString(sanitizeHeaderValue(header.value))
		msg.WriteString("\r\n")
	}
	msg.WriteString("\r\n")
	msg.WriteString(buildContactMailBody(contact))
	return msg.Bytes(), nil
}

func parseHeaderAddress(value string, errorCode string) (*mail.Address, error) {
	address, err := mail.ParseAddress(sanitizeHeaderValue(value))
	if err != nil {
		return nil, errors.New(errorCode)
	}
	return address, nil
}

func buildContactMailBody(contact models.Contact) string {
	return "Name: " + contact.Name + "\n" +
		"E-Mail: " + contact.Email + "\n" +
		"Telefon: " + contact.Phone + "\n\n" +
		"Nachricht:\n" + contact.Message
}

func sanitizeHeaderValue(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, ":", " ")
	return strings.Join(strings.Fields(value), " ")
}

func sanitizeMailError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, errSMTPNotConfigured) {
		return "smtp_not_configured"
	}
	return "smtp_send_failed"
}
