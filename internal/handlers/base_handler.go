package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"labbi-app/internal/models"
)

var templateDir = "internal/templates"

func SetTemplateDir(dir string) {
	if dir != "" {
		templateDir = dir
	}
}

func renderTemplate(w http.ResponseWriter, page string, data interface{}) error {
	basePath := filepath.Join(templateDir, "base.html")
	pagePath := filepath.Join(templateDir, page)
	tmpl, err := template.ParseFiles(basePath, pagePath)
	if err != nil {
		log.Printf("Fehler beim Parsen der Templates: %v", err)
		http.Error(w, "Fehler beim Laden der Seite", http.StatusInternalServerError)
		return err
	}

	err = tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Fehler beim Rendern von %s: %v", page, err)
		http.Error(w, "Fehler beim Rendern der Seite", http.StatusInternalServerError)
	}
	return err
}

func renderAdminTemplate(w http.ResponseWriter, page string, data interface{}) error {
	basePath := filepath.Join(templateDir, "admin_base.html")
	pagePath := filepath.Join(templateDir, page)

	funcMap := template.FuncMap{
		"join":          strings.Join,
		"contains":      containsString,
		"parentDogName": parentDogName,
	}

	tmpl, err := template.New("admin_base.html").Funcs(funcMap).ParseFiles(basePath, pagePath)
	if err != nil {
		log.Printf("Fehler beim Parsen der Admin-Templates (%s & %s): %v", basePath, pagePath, err)
		http.Error(w, "Fehler beim Laden der Seite", http.StatusInternalServerError)
		return err
	}

	err = tmpl.ExecuteTemplate(w, "admin_base.html", data)
	if err != nil {
		log.Printf("Fehler beim Rendern der Admin-Seite %s: %v", page, err)
		http.Error(w, "Fehler beim Rendern der Seite", http.StatusInternalServerError)
	}
	return err
}

func parentDogName(id string) string {
	dog := models.ParentDogByID(models.NormalizeParentDogID(id))
	if dog.Name == "" {
		return id
	}
	return dog.Name
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func readPublicFile(name string) ([]byte, error) {
	paths := []string{
		name,
		filepath.Join(filepath.Dir(templateDir), name),
		filepath.Join(filepath.Dir(filepath.Dir(templateDir)), name),
	}
	for _, path := range paths {
		content, err := os.ReadFile(filepath.Clean(path))
		if err == nil {
			return content, nil
		}
	}
	return nil, os.ErrNotExist
}
