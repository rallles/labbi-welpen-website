package router

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"labbi-app/internal/config"
	"labbi-app/internal/handlers"
	"labbi-app/internal/middleware"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func SetupRoutes(mux *http.ServeMux, driver neo4j.DriverWithContext, cfg config.Config) {
	// 1) Spezifische Seiten zuerst
	mux.HandleFunc("/about", handlers.AboutHandler)
	mux.HandleFunc("/dogs", handlers.DogsHandler)
	mux.HandleFunc("/puppies", handlers.PuppiesHandler)
	mux.HandleFunc("/list-puppies", handlers.ListPuppiesHandler)
	mux.HandleFunc("/healthz", handlers.HealthHandler)

	mux.HandleFunc("/contact", func(w http.ResponseWriter, r *http.Request) {
		handlers.ContactHandler(w, r, cfg)
	})
	mux.HandleFunc("/impressum", handlers.ImpressumHandler)
	mux.HandleFunc("/datenschutz", handlers.DatenschutzHandler)
	mux.HandleFunc("/robots.txt", handlers.RobotsHandler)
	mux.HandleFunc("/sitemap.xml", handlers.SitemapHandler)

	// Admin-Bereich: Formular anzeigen (GET) und verarbeiten (POST)
	//mux.HandleFunc("/admin/puppies/add", middleware.AuthMiddleware(handlers.AddPuppyHandler))
	// Admin-Dashboard (zeigt Erfolgs- oder Fehlermeldung an)
	mux.HandleFunc("/admin", middleware.AuthMiddleware(cfg, handlers.AdminDashboardHandler))
	mux.HandleFunc("/admin/puppies",
		middleware.AuthMiddleware(cfg, func(w http.ResponseWriter, r *http.Request) {
			handlers.ListPuppiesAdminHandler(w, r, driver)
		}))

	// Admin: Welpen löschen (POST), per BasicAuth geschützt
	mux.HandleFunc("/admin/puppies/delete",
		middleware.AuthMiddleware(cfg, func(w http.ResponseWriter, r *http.Request) {
			handlers.DeletePuppyHandler(w, r, driver)
		}),
	)

	// Admin: Welpen bearbeiten (GET/POST), per BasicAuth geschützt
	mux.HandleFunc("/admin/puppies/edit",
		middleware.AuthMiddleware(cfg, func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				handlers.EditPuppyFormHandler(w, r, driver) // Zeigt das Edit-Formular an
			case http.MethodPost:
				handlers.EditPuppySaveHandler(w, r, driver) // Speichert die Änderung
			default:
				http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
			}
		}),
	)

	// Admin: Formular anzeigen (GET) und verarbeiten (POST), per BasicAuth geschützt
	mux.HandleFunc("/admin/puppies/add",
		middleware.AuthMiddleware(cfg, func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodGet:
				handlers.AddPuppyFormHandler(w, r)
			case http.MethodPost:
				// Hier übergibst Du den bereits erstellten Driver
				handlers.AddPuppyHandler(w, r, driver, cfg)
			default:
				http.Error(w, "Methode nicht erlaubt", http.StatusMethodNotAllowed)
			}
		}),
	)

	// 2) Statische Dateien
	// 1) Arbeitsverzeichnis ermitteln
	// wir starten in cmd/, also eine Ebene hoch
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Arbeitsverzeichnis: %v", err)
	}
	projectRoot := filepath.Dir(wd)
	staticDir := filepath.Join(projectRoot, "static")
	// 2) Statische Dateien
	mux.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))),
	)
	mux.Handle("/uploads/",
		http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir))),
	)

	// 3) Nur exakt "/" → HomeHandler
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			handlers.HomeHandler(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
}
