package handlers

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
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
	contactRateLimitWindow = 10 * time.Minute
	contactRateLimitMax    = 5
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
	mu       sync.Mutex
	requests map[string][]time.Time
}

var defaultContactRateLimiter = &contactRateLimiter{requests: make(map[string][]time.Time)}

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
	if err := r.ParseForm(); err != nil {
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

	mailErr := sendContactMail(cfg, contact)
	contact.MailSent = mailErr == nil && smtpConfigured(cfg)
	contact.MailError = sanitizeMailError(mailErr)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := repository.NewContactRepository(driver).Create(ctx, contact); err != nil {
		log.Printf("Kontaktanfrage konnte nicht gespeichert werden: %v", err)
		renderContactForm(w, form, []string{"Ihre Anfrage konnte gerade nicht gespeichert werden. Bitte versuchen Sie es später erneut."})
		return
	}

	if mailErr != nil && smtpConfigured(cfg) {
		log.Printf("E-Mail-Benachrichtigung fehlgeschlagen: %v", mailErr)
	}

	renderTemplate(w, "contact_result.html", ContactResultData{
		Name:           contact.Name,
		Saved:          true,
		MailConfigured: smtpConfigured(cfg),
		MailSent:       contact.MailSent,
		MailFailed:     mailErr != nil && smtpConfigured(cfg),
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

func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if realIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); realIP != "" {
		return realIP
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func smtpConfigured(cfg config.Config) bool {
	return cfg.SMTPHost != "" && cfg.SMTPPort != "" && cfg.SMTPUser != "" &&
		cfg.SMTPPassword != "" && cfg.ContactMailTo != ""
}

// sendContactMail versendet eine Benachrichtigungs-E-Mail.
func sendContactMail(cfg config.Config, contact models.Contact) error {
	if !smtpConfigured(cfg) {
		return fmt.Errorf("smtp_not_configured")
	}

	auth := smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPassword, cfg.SMTPHost)
	subject := "Neue Kontaktanfrage von " + contact.Name
	body := "Name: " + contact.Name + "\n" +
		"E-Mail: " + contact.Email + "\n" +
		"Telefon: " + contact.Phone + "\n\n" +
		"Nachricht:\n" + contact.Message

	msgData := []byte("Subject: " + subject + "\r\n" +
		"\r\n" + body)

	return smtp.SendMail(cfg.SMTPHost+":"+cfg.SMTPPort, auth, cfg.SMTPUser, []string{cfg.ContactMailTo}, msgData)
}

func sanitizeMailError(err error) string {
	if err == nil {
		return ""
	}
	if err.Error() == "smtp_not_configured" {
		return "smtp_not_configured"
	}
	return "smtp_send_failed"
}
