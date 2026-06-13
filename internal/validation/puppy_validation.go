package validation

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"labbi-app/internal/models"
)

type PuppyForm struct {
	Name         string
	Geburtsdatum string
	Geschlecht   string
	Farbe        string
	Gewicht      string
	Charakter    string
	Geimpft      bool
	Gechippt     bool
	Entwurmt     bool
	Eltern       []string
	Notizen      string
}

func PuppyFormFromValues(values url.Values) PuppyForm {
	return PuppyForm{
		Name:         strings.TrimSpace(values.Get("name")),
		Geburtsdatum: strings.TrimSpace(values.Get("geburtsdatum")),
		Geschlecht:   strings.TrimSpace(values.Get("geschlecht")),
		Farbe:        strings.TrimSpace(values.Get("farbe")),
		Gewicht:      strings.TrimSpace(values.Get("gewicht")),
		Charakter:    strings.TrimSpace(values.Get("charakter")),
		Geimpft:      parseBool(values.Get("geimpft")),
		Gechippt:     parseBool(values.Get("gechippt")),
		Entwurmt:     parseBool(values.Get("entwurmt")),
		Eltern:       cleanParents(values["eltern"]),
		Notizen:      strings.TrimSpace(values.Get("notizen")),
	}
}

func ValidatePuppyForm(form PuppyForm) ([]string, float64) {
	var errs []string

	if form.Name == "" {
		errs = append(errs, "Name ist erforderlich.")
	} else if utf8.RuneCountInString(form.Name) > 80 {
		errs = append(errs, "Name darf maximal 80 Zeichen lang sein.")
	}

	if form.Geburtsdatum == "" {
		errs = append(errs, "Geburtsdatum ist erforderlich.")
	} else {
		birthdate, err := time.Parse("2006-01-02", form.Geburtsdatum)
		if err != nil {
			errs = append(errs, "Geburtsdatum muss im Format YYYY-MM-DD angegeben werden.")
		} else if birthdate.After(time.Now().Add(24 * time.Hour)) {
			errs = append(errs, "Geburtsdatum darf nicht in der Zukunft liegen.")
		} else if birthdate.Before(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
			errs = append(errs, "Geburtsdatum ist nicht plausibel.")
		}
	}

	if form.Geschlecht != "männlich" && form.Geschlecht != "weiblich" {
		errs = append(errs, "Geschlecht muss männlich oder weiblich sein.")
	}

	if !models.IstGueltigeFarbe(models.Fellfarbe(form.Farbe)) {
		errs = append(errs, "Fellfarbe ist ungültig.")
	}

	weight, err := strconv.ParseFloat(form.Gewicht, 64)
	if err != nil {
		errs = append(errs, "Gewicht muss eine Zahl sein.")
	} else if weight <= 0 || weight > 80 {
		errs = append(errs, "Gewicht muss größer 0 und plausibel sein.")
	}

	if utf8.RuneCountInString(form.Charakter) > 1000 {
		errs = append(errs, "Charakter darf maximal 1000 Zeichen lang sein.")
	}
	if utf8.RuneCountInString(form.Notizen) > 2000 {
		errs = append(errs, "Notizen dürfen maximal 2000 Zeichen lang sein.")
	}

	for _, parent := range form.Eltern {
		if !models.IstBekannterElternhund(parent) {
			errs = append(errs, fmt.Sprintf("Unbekannter Elternwert: %s.", parent))
		}
	}

	return errs, weight
}

func parseBool(value string) bool {
	return value == "true" || value == "on" || value == "1"
}

func cleanParents(values []string) []string {
	parents := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			parent := models.NormalizeParentDogID(strings.TrimSpace(item))
			if parent == "" || seen[parent] {
				continue
			}
			seen[parent] = true
			parents = append(parents, parent)
		}
	}
	return parents
}
