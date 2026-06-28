package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"labbi-app/internal/config"
	"labbi-app/internal/repository"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func DeletePuppyHandler(w http.ResponseWriter, r *http.Request, driver neo4j.DriverWithContext, cfg config.Config) {
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
		http.Redirect(w, r, "/admin/puppies?error=missing_id", http.StatusSeeOther)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	repo := repository.NewPuppyRepository(driver)
	puppy, err := repo.GetByID(ctx, id)
	if err != nil {
		log.Printf("Fehler beim Laden des zu loeschenden Welpen %q: %v", id, err)
		http.Redirect(w, r, "/admin/puppies?error=delete_failed", http.StatusSeeOther)
		return
	}

	if err := repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrPuppyNotFound) {
			log.Printf("Zu loeschender Welpe %q wurde nicht gefunden: %v", id, err)
		} else {
			log.Printf("Fehler beim Löschen des Welpen: %v", err)
		}
		http.Redirect(w, r, "/admin/puppies?error=delete_failed", http.StatusSeeOther)
		return
	}

	if err := removePuppyUploadFiles(cfg.UploadDir, puppy.Bilder); err != nil {
		log.Printf("Welpe %q wurde geloescht, aber Uploads konnten nicht vollstaendig entfernt werden: %v", id, err)
		http.Redirect(w, r, "/admin/puppies?success=deleted_with_upload_warning", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin/puppies?success=deleted", http.StatusSeeOther)
}

func removePuppyUploadFiles(uploadDir string, publicPaths []string) error {
	if strings.TrimSpace(uploadDir) == "" {
		return errors.New("upload directory is empty")
	}

	var cleanupErrors []error
	for _, publicPath := range publicPaths {
		if !strings.HasPrefix(publicPath, "/uploads/") || publicPath == "/uploads/" {
			continue
		}

		relativeName := strings.TrimPrefix(publicPath, "/uploads/")
		name := filepath.Base(publicPath)
		if name == "" || name == "." || name == string(filepath.Separator) ||
			relativeName != name || strings.ContainsAny(relativeName, `/\`) {
			continue
		}

		target := filepath.Join(uploadDir, name)
		if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
			cleanupErrors = append(cleanupErrors, fmt.Errorf("remove upload %q: %w", publicPath, err))
		}
	}
	return errors.Join(cleanupErrors...)
}
