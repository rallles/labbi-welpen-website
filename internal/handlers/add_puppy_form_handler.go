package handlers

import (
	"log"
	"net/http"
)

func AddPuppyFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	token, err := newCSRFToken()
	if err != nil {
		log.Printf("CSRF-Token konnte nicht erzeugt werden: %v", err)
		http.Error(w, "Serverfehler", http.StatusInternalServerError)
		return
	}

	if err := renderAdminTemplate(w, "admin/add_puppy.html", PuppyFormData{CSRFToken: token}); err != nil {
		log.Printf("Fehler beim Anzeigen des Add-Puppy-Formulars: %v", err)
	}
}
