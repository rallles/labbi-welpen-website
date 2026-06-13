package handlers

import (
	"net/http"
	"os"
	"path/filepath"
)

func DatenschutzHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}
	renderTemplate(w, "datenschutz.html", nil)
}

func RobotsHandler(w http.ResponseWriter, r *http.Request) {
	serveTextFile(w, r, "robots.txt", "text/plain; charset=utf-8")
}

func SitemapHandler(w http.ResponseWriter, r *http.Request) {
	serveTextFile(w, r, "sitemap.xml", "application/xml; charset=utf-8")
}

func serveTextFile(w http.ResponseWriter, r *http.Request, name string, contentType string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
		return
	}
	content, err := readPublicFile(name)
	if err != nil {
		http.Error(w, "Datei nicht gefunden", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(content)
}

func readPublicFile(name string) ([]byte, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	paths := []string{
		filepath.Join(wd, name),
		filepath.Join(filepath.Dir(wd), name),
	}
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err == nil {
			return content, nil
		}
	}
	return nil, os.ErrNotExist
}
