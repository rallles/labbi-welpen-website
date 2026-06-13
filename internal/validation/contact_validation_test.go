package validation

import "testing"

func TestValidateContactFormRejectsControlCharactersInHeaders(t *testing.T) {
	form := ContactForm{
		Name:    "Alice\r\nBcc: attacker@example.invalid",
		Email:   "alice@example.invalid",
		Phone:   "+49 123",
		Message: "Hallo zusammen",
	}

	errs := ValidateContactForm(form)
	if len(errs) == 0 {
		t.Fatal("ValidateContactForm() returned no errors, want control-character error")
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
	if len(errs) == 0 {
		t.Fatal("ValidateContactForm() returned no errors, want phone control-character error")
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
	if len(errs) == 0 {
		t.Fatal("ValidateContactForm() returned no errors, want message control-character error")
	}
}
