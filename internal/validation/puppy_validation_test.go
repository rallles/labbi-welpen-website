package validation

import (
	"net/url"
	"reflect"
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

func TestValidatePuppyFormTooLongName(t *testing.T) {
	form := validPuppyForm()
	form.Name = strings.Repeat("a", 81)

	errors, _ := ValidatePuppyForm(form)
	if !containsValidationError(errors, "Name darf maximal 80 Zeichen lang sein.") {
		t.Fatalf("expected name-length error, got %v", errors)
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

func TestValidatePuppyFormEmptyWeight(t *testing.T) {
	form := validPuppyForm()
	form.Gewicht = ""

	errors, _ := ValidatePuppyForm(form)
	if !containsValidationError(errors, "Gewicht muss eine Zahl sein.") {
		t.Fatalf("expected numeric-weight error, got %v", errors)
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

func TestValidatePuppyFormRejectsImplausiblyHighWeight(t *testing.T) {
	form := validPuppyForm()
	form.Gewicht = "80.1"

	errors, _ := ValidatePuppyForm(form)
	if !containsValidationError(errors, "Gewicht muss größer 0 und plausibel sein.") {
		t.Fatalf("expected implausible-weight error, got %v", errors)
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

func TestValidatePuppyFormTooLongNotes(t *testing.T) {
	form := validPuppyForm()
	form.Notizen = strings.Repeat("a", 2001)

	errors, _ := ValidatePuppyForm(form)
	if !containsValidationError(errors, "Notizen dürfen maximal 2000 Zeichen lang sein.") {
		t.Fatalf("expected notes-length error, got %v", errors)
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

func TestPuppyFormFromValuesNormalizesParentCase(t *testing.T) {
	form := PuppyFormFromValues(validPuppyValues(url.Values{
		"eltern": []string{"Gandalf", "Anna"},
	}))

	want := []string{"gandalf", "anna"}
	if !reflect.DeepEqual(form.Eltern, want) {
		t.Fatalf("PuppyFormFromValues().Eltern = %v, want %v", form.Eltern, want)
	}
}

func TestPuppyFormFromValuesRemovesDuplicateParents(t *testing.T) {
	form := PuppyFormFromValues(validPuppyValues(url.Values{
		"eltern": []string{"Gandalf, gandalf", "Anna", "anna"},
	}))

	want := []string{"gandalf", "anna"}
	if !reflect.DeepEqual(form.Eltern, want) {
		t.Fatalf("PuppyFormFromValues().Eltern = %v, want %v", form.Eltern, want)
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

func validPuppyValues(overrides url.Values) url.Values {
	values := url.Values{
		"name":         []string{"Balu"},
		"geburtsdatum": []string{"2024-01-10"},
		"geschlecht":   []string{"männlich"},
		"farbe":        []string{"schwarz"},
		"gewicht":      []string{"12.5"},
		"charakter":    []string{"freundlich"},
		"geimpft":      []string{"true"},
		"gechippt":     []string{"true"},
		"entwurmt":     []string{"true"},
		"eltern":       []string{"gandalf", "anna"},
		"notizen":      []string{"alles gut"},
	}
	for key, value := range overrides {
		values[key] = value
	}
	return values
}

func containsValidationError(errs []string, want string) bool {
	for _, err := range errs {
		if err == want {
			return true
		}
	}
	return false
}
