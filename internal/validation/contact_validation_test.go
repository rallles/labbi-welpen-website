package validation

import (
	"strings"
	"testing"
)

func TestValidateContactFormValidRequest(t *testing.T) {
	form := ContactForm{
		Name:    "Alice Example",
		Email:   "alice@example.invalid",
		Phone:   "+49 123",
		Message: "Hallo zusammen, ich habe Interesse am Wurf.",
	}

	if errs := ValidateContactForm(form); len(errs) > 0 {
		t.Fatalf("ValidateContactForm() errors = %v, want none", errs)
	}
}

func TestValidateContactFormInvalidEmail(t *testing.T) {
	form := ContactForm{
		Name:    "Alice Example",
		Email:   "alice@@example.invalid",
		Phone:   "+49 123",
		Message: "Hallo zusammen, ich habe Interesse.",
	}

	errors := ValidateContactForm(form)
	if !containsValidationError(errors, "Bitte geben Sie eine gültige E-Mail-Adresse ein.") {
		t.Fatalf("expected invalid-email error, got %v", errors)
	}
}

func TestValidateContactFormMessageTooShort(t *testing.T) {
	form := ContactForm{
		Name:    "Alice Example",
		Email:   "alice@example.invalid",
		Message: "zu kurz",
	}

	errors := ValidateContactForm(form)
	if !containsValidationError(errors, "Nachricht muss mindestens 10 Zeichen lang sein.") {
		t.Fatalf("expected short-message error, got %v", errors)
	}
}

func TestValidateContactFormMessageTooLong(t *testing.T) {
	form := ContactForm{
		Name:    "Alice Example",
		Email:   "alice@example.invalid",
		Message: strings.Repeat("a", 3001),
	}

	errors := ValidateContactForm(form)
	if !containsValidationError(errors, "Nachricht darf maximal 3000 Zeichen lang sein.") {
		t.Fatalf("expected long-message error, got %v", errors)
	}
}

func TestValidateContactFormNameWithControlCharacters(t *testing.T) {
	form := ContactForm{
		Name:    "Alice\r\nBcc: attacker@example.invalid",
		Email:   "alice@example.invalid",
		Phone:   "+49 123",
		Message: "Hallo zusammen",
	}

	errs := ValidateContactForm(form)
	if !containsValidationError(errs, "Name darf keine Steuerzeichen enthalten.") {
		t.Fatalf("expected control-character error in name, got %v", errs)
	}
}

func TestValidateContactFormRejectsControlCharactersInHeaders(t *testing.T) {
	form := ContactForm{
		Name:    "Alice\r\nBcc: attacker@example.invalid",
		Email:   "alice@example.invalid",
		Phone:   "+49 123",
		Message: "Hallo zusammen",
	}

	errs := ValidateContactForm(form)
	if !containsValidationError(errs, "Name darf keine Steuerzeichen enthalten.") {
		t.Fatalf("expected control-character error in name, got %v", errs)
	}
}

func TestValidateContactFormRejectsControlCharactersInPhone(t *testing.T) {
	form := ContactForm{
		Name:    "Alice",
		Email:   "alice@example.invalid",
		Phone:   "+49 123\x00",
		Message: "Hallo zusammen",
	}

	errs := ValidateContactForm(form)
	if !containsValidationError(errs, "Telefon darf keine Steuerzeichen enthalten.") {
		t.Fatalf("expected control-character error in phone, got %v", errs)
	}
}

func TestValidateContactFormAllowsMessageLineBreaks(t *testing.T) {
	form := ContactForm{
		Name:    "Alice",
		Email:   "alice@example.invalid",
		Phone:   "+49 123",
		Message: "Hallo,\nbitte melden.\nDanke.",
	}

	if errs := ValidateContactForm(form); len(errs) > 0 {
		t.Fatalf("ValidateContactForm() errors = %v, want none", errs)
	}
}

func TestValidateContactFormRejectsExtremeMessageControls(t *testing.T) {
	form := ContactForm{
		Name:    "Alice",
		Email:   "alice@example.invalid",
		Message: "Hallo zusammen\x00",
	}

	errs := ValidateContactForm(form)
	if !containsValidationError(errs, "Nachricht enthält ungültige Steuerzeichen.") {
		t.Fatalf("expected control-character error in message, got %v", errs)
	}
}
