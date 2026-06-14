package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"labbi-app/internal/repository"
	"labbi-app/internal/validation"

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

	renderEditPuppyForm(w, id, puppyFormFromModel(puppy), nil)
}

func EditPuppySaveHandler(w http.ResponseWriter, r *http.Request, driver neo4j.DriverWithContext) {
	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ungültige Eingaben", http.StatusBadRequest)
		return
	}
	if !consumeCSRFToken(r.FormValue("csrf_token")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "ID fehlt", http.StatusBadRequest)
		return
	}

	form := validation.PuppyFormFromValues(r.Form)
	errs, weight := validation.ValidatePuppyForm(form)
	if len(errs) > 0 {
		renderEditPuppyForm(w, id, form, errs)
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
		renderEditPuppyForm(w, id, form, []string{"Der Welpe konnte nicht geladen werden. Bitte versuchen Sie es erneut."})
		return
	}

	puppy := puppyFromForm(id, form, weight, existing.Bilder)
	if err := repo.Update(ctx, puppy); err != nil {
		log.Printf("Fehler beim Speichern des Welpen: %v", err)
		renderEditPuppyForm(w, id, form, []string{"Der Welpe konnte nicht gespeichert werden. Bitte versuchen Sie es erneut."})
		return
	}

	http.Redirect(w, r, "/admin/puppies?success=updated", http.StatusSeeOther)
}

func renderEditPuppyForm(w http.ResponseWriter, id string, form validation.PuppyForm, errors []string) {
	token, err := newCSRFToken()
	if err != nil {
		log.Printf("CSRF-Token konnte nicht erzeugt werden: %v", err)
		http.Error(w, "Serverfehler", http.StatusInternalServerError)
		return
	}
	if err := renderAdminTemplate(w, "admin/admin_puppies_edit.html", PuppyFormData{ID: id, CSRFToken: token, Errors: errors, Form: form}); err != nil {
		log.Printf("Fehler beim Anzeigen des Edit-Formulars: %v", err)
	}
}
