package validation

import (
	"net/mail"
	"net/url"
	"strings"
	"unicode/utf8"
)

type ContactForm struct {
	Name    string
	Email   string
	Phone   string
	Message string
	Website string
}

func ContactFormFromValues(values url.Values) ContactForm {
	return ContactForm{
		Name:    strings.TrimSpace(values.Get("name")),
		Email:   strings.TrimSpace(values.Get("email")),
		Phone:   strings.TrimSpace(values.Get("phone")),
		Message: strings.TrimSpace(values.Get("message")),
		Website: strings.TrimSpace(values.Get("website")),
	}
}

func ValidateContactForm(form ContactForm) []string {
	var errs []string
	if form.Name == "" {
		errs = append(errs, "Name ist erforderlich.")
	} else if utf8.RuneCountInString(form.Name) > 120 {
		errs = append(errs, "Name darf maximal 120 Zeichen lang sein.")
	}

	if form.Email == "" {
		errs = append(errs, "E-Mail ist erforderlich.")
	} else if _, err := mail.ParseAddress(form.Email); err != nil {
		errs = append(errs, "Bitte geben Sie eine gültige E-Mail-Adresse ein.")
	}

	if utf8.RuneCountInString(form.Phone) > 40 {
		errs = append(errs, "Telefon darf maximal 40 Zeichen lang sein.")
	}

	messageLen := utf8.RuneCountInString(form.Message)
	if form.Message == "" {
		errs = append(errs, "Nachricht ist erforderlich.")
	} else if messageLen < 10 {
		errs = append(errs, "Nachricht muss mindestens 10 Zeichen lang sein.")
	} else if messageLen > 3000 {
		errs = append(errs, "Nachricht darf maximal 3000 Zeichen lang sein.")
	}
	return errs
}
