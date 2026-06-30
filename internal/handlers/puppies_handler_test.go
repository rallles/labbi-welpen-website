package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"labbi-app/internal/models"
)

func TestPuppiesTemplateRendersAllDataStates(t *testing.T) {
	originalTemplateDir := templateDir
	SetTemplateDir("../templates")
	t.Cleanup(func() { templateDir = originalTemplateDir })

	t.Run("entry", func(t *testing.T) {
		response := httptest.NewRecorder()
		page := PuppiesPage{Puppies: []models.Puppy{{
			Name:         "Luna",
			Geburtsdatum: "2026-05-01",
			Geschlecht:   "weiblich",
			Farbe:        models.FarbeSchwarz,
			Gewicht:      4.2,
			Bilder:       []string{"/uploads/luna.jpg"},
		}}}

		if err := renderTemplate(response, "puppies.html", page); err != nil {
			t.Fatalf("renderTemplate() error = %v", err)
		}
		body := response.Body.String()
		for _, expected := range []string{"Luna", "2026-05-01", "/uploads/luna.jpg"} {
			if !strings.Contains(body, expected) {
				t.Errorf("rendered page does not contain %q", expected)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		response := httptest.NewRecorder()
		if err := renderTemplate(response, "puppies.html", PuppiesPage{}); err != nil {
			t.Fatalf("renderTemplate() error = %v", err)
		}
		if !strings.Contains(response.Body.String(), "Aktuell sind keine Welpen eingetragen") {
			t.Error("rendered page does not contain the empty-state message")
		}
	})

	t.Run("load error", func(t *testing.T) {
		response := httptest.NewRecorder()
		page := PuppiesPage{LoadError: puppiesLoadErrorMessage}
		if err := renderTemplate(response, "puppies.html", page); err != nil {
			t.Fatalf("renderTemplate() error = %v", err)
		}

		body := response.Body.String()
		if !strings.Contains(body, page.LoadError) {
			t.Error("rendered page does not contain the load-error message")
		}
		if strings.Contains(body, "Aktuell sind keine Welpen eingetragen") {
			t.Error("load-error state incorrectly claims that no puppies exist")
		}
		for _, fixedContent := range []string{"Feste Galerie", "/static/images/generated/welpe_001_-480.jpg"} {
			if !strings.Contains(body, fixedContent) {
				t.Errorf("load-error state does not contain fixed content %q", fixedContent)
			}
		}
	})
}
