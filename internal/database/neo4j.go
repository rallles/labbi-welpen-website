package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"labbi-app/internal/config"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// NewNeo4jDriver erstellt einen Neo4j-Driver basierend auf der Konfiguration.
func NewNeo4jDriver(cfg config.Config) (neo4j.DriverWithContext, error) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := EnsureConstraints(ctx, driver); err != nil {
		_ = driver.Close(context.Background())
		return nil, err
	}

	return driver, nil
}

func EnsureConstraints(ctx context.Context, driver neo4j.DriverWithContext) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	result, err := session.Run(ctx,
		"CREATE CONSTRAINT puppy_id IF NOT EXISTS FOR (p:Puppy) REQUIRE p.id IS UNIQUE",
		nil,
	)
	if err != nil {
		return fmt.Errorf("create puppy id constraint: %w", err)
	}
	if _, err := result.Consume(ctx); err != nil {
		return fmt.Errorf("consume puppy id constraint result: %w", err)
	}
	return nil
}

// DefaultSessionConfig liefert die Standard-Session-Einstellungen zurück.
func DefaultSessionConfig() neo4j.SessionConfig {
	return neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	}
}
