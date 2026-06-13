package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"labbi-app/internal/models"
	"labbi-app/internal/repository"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func EditPuppyFormHandler(w http.ResponseWriter, r *http.Request, driver neo4j.DriverWithContext) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID fehlt", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	puppy, err := repository.NewPuppyRepository(driver).GetByID(ctx, id)
	if errors.Is(err, repository.ErrPuppyNotFound) {
		http.Error(w, "Welpe nicht gefunden", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Fehler beim Laden des Welpen: %v", err)
		http.Error(w, "Fehler beim Laden des Welpen", http.StatusInternalServerError)
		return
	}

	if err := renderAdminTemplate(w, "admin/admin_puppies_edit.html", puppy); err != nil {
		http.Error(w, "Fehler beim Anzeigen des Edit-Formulars", http.StatusInternalServerError)
	}
}

func EditPuppySaveHandler(w http.ResponseWriter, r *http.Request, driver neo4j.DriverWithContext) {
	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "ID fehlt", http.StatusBadRequest)
		return
	}

	weight, err := strconv.ParseFloat(r.FormValue("gewicht"), 64)
	if err != nil {
		http.Error(w, "Ungültiges Gewicht", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	repo := repository.NewPuppyRepository(driver)
	existing, err := repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrPuppyNotFound) {
		http.Error(w, "Welpe nicht gefunden", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Fehler beim Laden des Welpen: %v", err)
		http.Error(w, "Fehler beim Laden des Welpen", http.StatusInternalServerError)
		return
	}

	parents := strings.Split(r.FormValue("eltern"), ",")
	for i := range parents {
		parents[i] = strings.TrimSpace(parents[i])
	}

	puppy := models.Puppy{
		ID:           id,
		Name:         r.FormValue("name"),
		Geburtsdatum: r.FormValue("geburtsdatum"),
		Geschlecht:   r.FormValue("geschlecht"),
		Farbe:        models.Fellfarbe(r.FormValue("farbe")),
		Gewicht:      weight,
		Charakter:    r.FormValue("charakter"),
		Geimpft:      r.FormValue("geimpft") == "on",
		Gechippt:     r.FormValue("gechippt") == "on",
		Entwurmt:     r.FormValue("entwurmt") == "on",
		Eltern:       parents,
		Notizen:      r.FormValue("notizen"),
		Bilder:       existing.Bilder,
	}

	if err := repo.Update(ctx, puppy); err != nil {
		log.Printf("Fehler beim Speichern des Welpen: %v", err)
		http.Error(w, "Fehler beim Speichern", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/puppies", http.StatusSeeOther)
}
