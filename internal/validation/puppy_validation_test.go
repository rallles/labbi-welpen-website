package validation

import (
	"strings"
	"testing"
	"time"
)

func TestValidatePuppyFormValid(t *testing.T) {
	form := validPuppyForm()

	errors, weight := ValidatePuppyForm(form)
	if len(errors) != 0 {
		t.Fatalf("ValidatePuppyForm() errors = %v, want none", errors)
	}
	if weight != 12.5 {
		t.Fatalf("ValidatePuppyForm() weight = %v, want 12.5", weight)
	}
}

func TestValidatePuppyFormEmptyName(t *testing.T) {
	form := validPuppyForm()
	form.Name = ""

	errors, _ := ValidatePuppyForm(form)
	if !containsValidationError(errors, "Name ist erforderlich.") {
		t.Fatalf("expected name-required error, got %v", errors)
	}
}

func TestValidatePuppyFormInvalidBirthdateFormat(t *testing.T) {
	form := validPuppyForm()
	form.Geburtsdatum = "2026-99-01"

	errors, _ := ValidatePuppyForm(form)
	if !containsValidationError(errors, "Geburtsdatum muss im Format YYYY-MM-DD angegeben werden.") {
		t.Fatalf("expected invalid-date-format error, got %v", errors)
	}
}

func TestValidatePuppyFormFutureBirthdate(t *testing.T) {
	form := validPuppyForm()
	form.Geburtsdatum = time.Now().AddDate(0, 0, 3).Format("2006-01-02")

	errors, _ := ValidatePuppyForm(form)
	if !containsValidationError(errors, "Geburtsdatum darf nicht in der Zukunft liegen.") {
		t.Fatalf("expected future-date error, got %v", errors)
	}
}

func TestValidatePuppyFormInvalidGender(t *testing.T) {
	form := validPuppyForm()
	form.Geschlecht = "divers"

	errors, _ := ValidatePuppyForm(form)
	if !containsValidationError(errors, "Geschlecht muss männlich oder weiblich sein.") {
		t.Fatalf("expected invalid-gender error, got %v", errors)
	}
}

func TestValidatePuppyFormInvalidCoatColor(t *testing.T) {
	form := validPuppyForm()
	form.Farbe = "blau"

	errors, _ := ValidatePuppyForm(form)
	if !containsValidationError(errors, "Fellfarbe ist ungültig.") {
		t.Fatalf("expected invalid-color error, got %v", errors)
	}
}

func TestValidatePuppyFormInvalidWeight(t *testing.T) {
	form := validPuppyForm()
	form.Gewicht = "0"

	errors, _ := ValidatePuppyForm(form)
	if !containsValidationError(errors, "Gewicht muss größer 0 und plausibel sein.") {
		t.Fatalf("expected invalid-weight error, got %v", errors)
	}
}

func TestValidatePuppyFormTooLongCharacter(t *testing.T) {
	form := validPuppyForm()
	form.Charakter = strings.Repeat("a", 1001)

	errors, _ := ValidatePuppyForm(form)
	if !containsValidationError(errors, "Charakter darf maximal 1000 Zeichen lang sein.") {
		t.Fatalf("expected character-length error, got %v", errors)
	}
}

func TestValidatePuppyFormUnknownParent(t *testing.T) {
	form := validPuppyForm()
	form.Eltern = []string{"unknown-parent"}

	errors, _ := ValidatePuppyForm(form)
	if !containsValidationError(errors, "Unbekannter Elternwert: unknown-parent.") {
		t.Fatalf("expected unknown-parent error, got %v", errors)
	}
}

func validPuppyForm() PuppyForm {
	return PuppyForm{
		Name:         "Balu",
		Geburtsdatum: "2024-01-10",
		Geschlecht:   "männlich",
		Farbe:        "schwarz",
		Gewicht:      "12.5",
		Charakter:    "freundlich",
		Geimpft:      true,
		Gechippt:     true,
		Entwurmt:     true,
		Eltern:       []string{"gandalf", "anna"},
		Notizen:      "alles gut",
	}
}

func containsValidationError(errs []string, want string) bool {
	for _, err := range errs {
		if err == want {
			return true
		}
	}
	return false
}