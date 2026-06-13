package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"labbi-app/internal/config"
)

// AuthMiddleware schützt Admin-Routen mit Basic Auth
func AuthMiddleware(cfg config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || cfg.AdminUser == "" || cfg.AdminPassword == "" ||
			!constantTimeEqual(user, cfg.AdminUser) ||
			!constantTimeEqual(pass, cfg.AdminPassword) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Admin Bereich"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func constantTimeEqual(a, b string) bool {
	aHash := sha256.Sum256([]byte(a))
	bHash := sha256.Sum256([]byte(b))

	return subtle.ConstantTimeCompare(aHash[:], bHash[:]) == 1
}
