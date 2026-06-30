package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"labbi-app/internal/models"
	"labbi-app/internal/repository"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type PuppiesPage struct {
	Puppies   []models.Puppy
	LoadError string
}

const puppiesLoadErrorMessage = "Die aktuellen Welpendaten konnten gerade nicht geladen werden."

// MakePuppiesHandler erstellt die oeffentliche, aus Neo4j gespeiste Welpenseite.
func MakePuppiesHandler(driver neo4j.DriverWithContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		page := PuppiesPage{}
		puppies, err := repository.NewPuppyRepository(driver).List(ctx)
		if err != nil {
			log.Printf("Fehler beim Abfragen der oeffentlichen Welpenliste: %v", err)
			page.LoadError = puppiesLoadErrorMessage
		} else {
			page.Puppies = puppies
		}

		if err := renderTemplate(w, "puppies.html", page); err != nil {
			log.Printf("Fehler beim Rendern der oeffentlichen Welpenliste: %v", err)
		}
	}
}
