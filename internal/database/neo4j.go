package database

import (
	"fmt"
	"log"

	"labbi-app/internal/config"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// NewNeo4jDriver erstellt einen Neo4j-Driver basierend auf der Konfiguration.
func NewNeo4jDriver(cfg config.Config) (neo4j.DriverWithContext, error) {
	// Prüfen, ob URI gesetzt ist
	if cfg.Neo4jUri == "" {
		log.Println("WARN: Neo4j URI ist leer, verwende Default neo4j://localhost:7687")
		cfg.Neo4jUri = "neo4j://localhost:7687"
	}
	if cfg.Neo4jUser == "" || cfg.Neo4jPassword == "" {
		return nil, fmt.Errorf("neo4j credentials are not configured")
	}

	driver, err := neo4j.NewDriverWithContext(
		cfg.Neo4jUri,
		neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""),
	)
	if err != nil {
		return nil, err
	}
	return driver, nil
}

// DefaultSessionConfig liefert die Standard-Session-Einstellungen zurück.
func DefaultSessionConfig() neo4j.SessionConfig {
	return neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	}
}
