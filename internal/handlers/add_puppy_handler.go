package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"labbi-app/internal/config"
	"labbi-app/internal/repository"
	"labbi-app/internal/validation"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const (
	maxUploadImages    = 10
	maxUploadFileSize  = 5 << 20
	maxUploadTotalSize = 25 << 20
	multipartMemory    = 8 << 20
)

var errInvalidImageType = errors.New("invalid image type")

// AddPuppyHandler verarbeitet das Admin-Formular (POST) und speichert den neuen Welpen in Neo4j.
func AddPuppyHandler(w http.ResponseWriter, r *http.Request, driver neo4j.DriverWithContext, cfg config.Config) {
	if r.Method != http.MethodPost {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadTotalSize)
	if err := r.ParseMultipartForm(multipartMemory); err != nil {
		log.Printf("Fehler beim Lesen des Formulars: %v", err)
		status := http.StatusBadRequest
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			status = http.StatusRequestEntityTooLarge
		}
		http.Error(w, "Fehler beim Lesen des Formulars", status)
		return
	}

	if !consumeCSRFToken(r.FormValue("csrf_token")) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	form := validation.PuppyFormFromValues(r.MultipartForm.Value)
	errs, weight := validation.ValidatePuppyForm(form)
	if len(errs) > 0 {
		renderAddPuppyForm(w, form, errs)
		return
	}

	imagePaths, err := saveUploadedImages(r.MultipartForm.File["images"], cfg.UploadDir)
	if err != nil {
		log.Printf("Fehler beim Speichern der Bilder: %v", err)
		renderAddPuppyForm(w, form, []string{"Bilder konnten nicht verarbeitet werden. Bitte laden Sie JPEG- oder PNG-Dateien innerhalb der Größenlimits hoch."})
		return
	}

	puppy := puppyFromForm(uuid.NewString(), form, weight, imagePaths)

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	repo := repository.NewPuppyRepository(driver)
	if err := repo.Create(ctx, puppy); err != nil {
		cleanupUploadedImages(cfg.UploadDir, imagePaths)
		log.Printf("Fehler beim Speichern in Neo4j: %v", err)
		renderAddPuppyForm(w, form, []string{"Der Welpe konnte nicht gespeichert werden. Bitte versuchen Sie es erneut."})
		return
	}

	http.Redirect(w, r, "/admin?success=true", http.StatusSeeOther)
}

func renderAddPuppyForm(w http.ResponseWriter, form validation.PuppyForm, errors []string) {
	token, err := newCSRFToken()
	if err != nil {
		log.Printf("CSRF-Token konnte nicht erzeugt werden: %v", err)
		http.Error(w, "Serverfehler", http.StatusInternalServerError)
		return
	}
	if err := renderAdminTemplate(w, "admin/add_puppy.html", PuppyFormData{CSRFToken: token, Errors: errors, Form: form}); err != nil {
		log.Printf("Fehler beim Rendern des Add-Puppy-Formulars: %v", err)
	}
}

// saveUploadedImages speichert alle hochgeladenen Dateien und liefert öffentliche relative Pfade zurück.
func saveUploadedImages(files []*multipart.FileHeader, uploadDir string) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}
	if len(files) > maxUploadImages {
		return nil, fmt.Errorf("too many images: max %d", maxUploadImages)
	}
	var totalSize int64
	for _, fh := range files {
		totalSize += fh.Size
	}
	if totalSize > maxUploadTotalSize {
		return nil, fmt.Errorf("uploaded images exceed total limit of %d bytes", maxUploadTotalSize)
	}

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(files))
	for _, fh := range files {
		path, err := saveUploadedImage(fh, uploadDir)
		if err != nil {
			cleanupUploadedImages(uploadDir, paths)
			return nil, err
		}
		paths = append(paths, path)
	}
	return paths, nil
}

func saveUploadedImage(fh *multipart.FileHeader, uploadDir string) (string, error) {
	if fh.Size > maxUploadFileSize {
		return "", fmt.Errorf("image %q exceeds %d bytes", fh.Filename, maxUploadFileSize)
	}

	file, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	limited := io.LimitReader(file, maxUploadFileSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", err
	}
	if int64(len(data)) > maxUploadFileSize {
		return "", fmt.Errorf("image %q exceeds %d bytes", fh.Filename, maxUploadFileSize)
	}

	ext, err := validateImage(data)
	if err != nil {
		return "", err
	}

	name := uuid.NewString() + ext
	target := filepath.Join(uploadDir, name)
	if err := os.WriteFile(target, data, 0644); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}

func validateImage(data []byte) (string, error) {
	contentType := http.DetectContentType(data)
	switch contentType {
	case "image/jpeg":
		if _, err := jpeg.DecodeConfig(bytes.NewReader(data)); err != nil {
			return "", fmt.Errorf("decode jpeg: %w", err)
		}
		return ".jpg", nil
	case "image/png":
		if _, err := png.DecodeConfig(bytes.NewReader(data)); err != nil {
			return "", fmt.Errorf("decode png: %w", err)
		}
		return ".png", nil
	default:
		return "", errInvalidImageType
	}
}

func cleanupUploadedImages(uploadDir string, paths []string) {
	for _, publicPath := range paths {
		name := filepath.Base(publicPath)
		if name == "." || name == string(filepath.Separator) {
			continue
		}
		if err := os.Remove(filepath.Join(uploadDir, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Printf("Fehler beim Aufräumen von Upload %s: %v", publicPath, err)
		}
	}
}

func uploadErrorStatus(err error) int {
	if errors.Is(err, errInvalidImageType) || strings.Contains(err.Error(), "exceeds") || strings.Contains(err.Error(), "too many images") || strings.Contains(err.Error(), "total limit") {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
