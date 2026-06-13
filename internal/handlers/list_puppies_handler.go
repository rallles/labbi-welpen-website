package handlers

import "net/http"

func ListPuppiesHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/puppies", http.StatusMovedPermanently)
}
