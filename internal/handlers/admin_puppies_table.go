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

	if err := renderAdminTemplate(w, "admin/admin_puppies_table.html", puppies); err != nil {
		log.Printf("Fehler beim Rendern der Welpen-Liste: %v", err)
		http.Error(w, "Interner Fehler beim Rendern", http.StatusInternalServerError)
	}
}
