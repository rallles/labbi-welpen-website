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

	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "ID fehlt", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if err := repository.NewPuppyRepository(driver).Delete(ctx, id); err != nil {
		log.Printf("Fehler beim Löschen des Welpen: %v", err)
		http.Error(w, "Fehler beim Löschen des Welpen", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/puppies", http.StatusSeeOther)
}
