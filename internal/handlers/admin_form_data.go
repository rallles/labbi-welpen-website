package handlers

import (
	"fmt"
	"strings"

	"labbi-app/internal/models"
	"labbi-app/internal/security"
	"labbi-app/internal/validation"
)

type PuppyFormData struct {
	ID        string
	CSRFToken string
	Errors    []string
	Form      validation.PuppyForm
}

type AdminPuppiesPage struct {
	CSRFToken string
	Error     string
	Success   string
	Puppies   []models.Puppy
}

func newCSRFToken() (string, error) {
	return security.DefaultCSRF.Generate()
}

func consumeCSRFToken(token string) bool {
	return security.DefaultCSRF.Consume(token)
}

func puppyFormFromModel(puppy models.Puppy) validation.PuppyForm {
	return validation.PuppyForm{
		Name:         puppy.Name,
		Geburtsdatum: puppy.Geburtsdatum,
		Geschlecht:   puppy.Geschlecht,
		Farbe:        string(puppy.Farbe),
		Gewicht:      formatWeight(puppy.Gewicht),
		Charakter:    puppy.Charakter,
		Geimpft:      puppy.Geimpft,
		Gechippt:     puppy.Gechippt,
		Entwurmt:     puppy.Entwurmt,
		Eltern:       puppy.Eltern,
		Notizen:      puppy.Notizen,
	}
}

func puppyFromForm(id string, form validation.PuppyForm, weight float64, images []string) models.Puppy {
	return models.Puppy{
		ID:           id,
		Name:         form.Name,
		Geburtsdatum: form.Geburtsdatum,
		Geschlecht:   form.Geschlecht,
		Farbe:        models.Fellfarbe(form.Farbe),
		Gewicht:      weight,
		Charakter:    form.Charakter,
		Geimpft:      form.Geimpft,
		Gechippt:     form.Gechippt,
		Entwurmt:     form.Entwurmt,
		Eltern:       form.Eltern,
		Notizen:      form.Notizen,
		Bilder:       images,
	}
}

func formatWeight(weight float64) string {
	if weight == 0 {
		return ""
	}
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", weight), "0"), ".")
}
