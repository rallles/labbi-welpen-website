package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"labbi-app/internal/repository"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func DeletePuppyHandler(w http.ResponseWriter, r *http.Request, driver neo4j.DriverWithContext) {
	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Ungültige Eingaben", http.StatusBadRequest)
		return
	}
	if !validCSRFToken(r.FormValue("csrf_token")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		http.Redirect(w, r, "/admin/puppies?error=missing_id", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := repository.NewPuppyRepository(driver).Delete(ctx, id); err != nil {
		log.Printf("Fehler beim Löschen des Welpen: %v", err)
		http.Redirect(w, r, "/admin/puppies?error=delete_failed", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/puppies?success=deleted", http.StatusSeeOther)
}
