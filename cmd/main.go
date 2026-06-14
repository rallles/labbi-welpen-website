package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"labbi-app/internal/config"
	"labbi-app/internal/database"
	"labbi-app/internal/router"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("Warnung: .env konnte nicht geladen werden: %v", err)
	}

	cfg := config.LoadConfig()
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = ":8080"
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Ungültige Konfiguration: %v", err)
	}

	driver, err := database.NewNeo4jDriver(cfg)
	if err != nil {
		log.Fatalf("Neo4j-Driver konnte nicht initialisiert werden: %v", err)
	}
	defer driver.Close(context.Background())

	mux := http.NewServeMux()
	router.SetupRoutes(mux, driver, cfg)

	server := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Labbi-App läuft auf %s", cfg.ServerAddress)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server konnte nicht gestartet werden: %v", err)
	}
}
