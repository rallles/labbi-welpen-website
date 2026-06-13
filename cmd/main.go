package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"labbi-app/internal/config"
	"labbi-app/internal/database"
	"labbi-app/internal/router"

	"github.com/joho/godotenv"
)

func main() {

	//utils.InitAdminTemplates()

	// 1) Arbeitsverzeichnis ausgeben
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Arbeitsverzeichnis konnte nicht ermittelt werden: %v", err)
	}
	log.Printf("Arbeitsverzeichnis: %s", wd)

	// 2) .env laden — passe den Pfad an, falls nötig:
	envPath := filepath.Join(wd, "..", ".env") // eine Ebene höher
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("Warnung: .env nicht in %s gefunden (%v), versuche wd/.env", envPath, err)
		if err2 := godotenv.Load(); err2 != nil {
			log.Printf("Warnung: .env auch nicht in %s gefunden (%v)", wd, err2)
		}
	}

	//-------------------------------------------------------------
	// .env laden – ignoriert Fehler, wenn keine .env vorhanden ist
	//_ = godotenv.Load()
	//--------------------------------------------------------------

	// 1. Konfiguration laden
	cfg := config.LoadConfig()
	log.Printf("Konfig geladen: ServerAddress=%q Neo4jURI=%q", cfg.ServerAddress, cfg.Neo4jUri)
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = ":8080" // Standardport
	}

	// 2. Neo4j-Driver initialisieren
	driver, err := database.NewNeo4jDriver(cfg)
	if err != nil {
		log.Fatalf("Neo4j-Driver konnte nicht initialisiert werden: %v", err)
	}
	defer driver.Close(context.Background())

	// 3. ServeMux und Routing aufsetzen
	mux := http.NewServeMux()
	router.SetupRoutes(mux, driver, cfg)

	// 4. Server starten
	log.Printf("Labbi-App läuft auf %s", cfg.ServerAddress)
	err = http.ListenAndServe(cfg.ServerAddress, mux)
	if err != nil {
		log.Fatalf("Server konnte nicht gestartet werden: %v", err)
	}
}
