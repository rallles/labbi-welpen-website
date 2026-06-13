package handlers

import (
	"log"
	"net/http"
)

// PuppiesHandler rendert die bewusst statische Welpen-Seite.
func PuppiesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Welpen-Seite aufgerufen")
	renderTemplate(w, "puppies.html", nil)
}
