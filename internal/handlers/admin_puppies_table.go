package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"labbi-app/internal/repository"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func ListPuppiesAdminHandler(w http.ResponseWriter, r *http.Request, driver neo4j.DriverWithContext) {
	if r.Method != http.MethodGet {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	puppies, err := repository.NewPuppyRepository(driver).List(ctx)
	if err != nil {
		log.Printf("Fehler beim Abfragen der Welpen: %v", err)
		http.Error(w, "Fehler beim Laden der Daten", http.StatusInternalServerError)
		return
	}

	token, err := newCSRFToken()
	if err != nil {
		log.Printf("CSRF-Token konnte nicht erzeugt werden: %v", err)
		http.Error(w, "Serverfehler", http.StatusInternalServerError)
		return
	}

	page := AdminPuppiesPage{
		CSRFToken: token,
		Error:     adminErrorMessage(r.URL.Query().Get("error")),
		Success:   adminSuccessMessage(r.URL.Query().Get("success")),
		Puppies:   puppies,
	}
	if err := renderAdminTemplate(w, "admin/admin_puppies_table.html", page); err != nil {
		log.Printf("Fehler beim Rendern der Welpen-Liste: %v", err)
		http.Error(w, "Interner Fehler beim Rendern", http.StatusInternalServerError)
	}
}

func adminErrorMessage(code string) string {
	switch code {
	case "missing_id":
		return "Der Welpe konnte nicht gelöscht werden: ID fehlt."
	case "delete_failed":
		return "Der Welpe konnte nicht gelöscht werden. Bitte versuchen Sie es erneut."
	default:
		return ""
	}
}

func adminSuccessMessage(code string) string {
	switch code {
	case "deleted":
		return "Der Welpe wurde gelöscht."
	case "updated":
		return "Der Welpe wurde gespeichert."
	default:
		return ""
	}
}
