package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"labbi-app/internal/config"
	"labbi-app/internal/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// NewNeo4jDriver erstellt einen Neo4j-Driver basierend auf der Konfiguration.
func NewNeo4jDriver(cfg config.Config) (neo4j.DriverWithContext, error) {
	if strings.TrimSpace(cfg.Neo4jUri) == "" {
		return nil, fmt.Errorf("neo4j URI is not configured")
	}
	if strings.TrimSpace(cfg.Neo4jUser) == "" || strings.TrimSpace(cfg.Neo4jPassword) == "" {
		return nil, fmt.Errorf("neo4j credentials are not configured")
	}

	driver, err := neo4j.NewDriverWithContext(
		cfg.Neo4jUri,
		neo4j.BasicAuth(cfg.Neo4jUser, cfg.Neo4jPassword, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("create neo4j driver: %w", err)
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

	constraints := []struct {
		name  string
		query string
	}{
		{
			name:  "puppy id constraint",
			query: "CREATE CONSTRAINT puppy_id IF NOT EXISTS FOR (p:Puppy) REQUIRE p.id IS UNIQUE",
		},
		{
			name:  "dog id constraint",
			query: "CREATE CONSTRAINT dog_id IF NOT EXISTS FOR (d:Dog) REQUIRE d.id IS UNIQUE",
		},
		{
			name:  "contact id constraint",
			query: "CREATE CONSTRAINT contact_id IF NOT EXISTS FOR (c:Contact) REQUIRE c.id IS UNIQUE",
		},
	}

	for _, constraint := range constraints {
		result, err := session.Run(ctx, constraint.query, nil)
		if err != nil {
			return fmt.Errorf("create %s: %w", constraint.name, err)
		}
		if _, err := result.Consume(ctx); err != nil {
			return fmt.Errorf("consume %s result: %w", constraint.name, err)
		}
	}

	for _, dog := range models.ParentDogs {
		result, err := session.Run(ctx, `
			MERGE (d:Dog {id: $id})
			SET d.name = $name,
				d.gender = $gender,
				d.role = $role`,
			map[string]any{
				"id":     dog.ID,
				"name":   dog.Name,
				"gender": dog.Gender,
				"role":   dog.Role,
			})
		if err != nil {
			return fmt.Errorf("seed parent dog %s: %w", dog.ID, err)
		}
		if _, err := result.Consume(ctx); err != nil {
			return fmt.Errorf("consume parent dog %s seed result: %w", dog.ID, err)
		}
	}
	return nil
}

// DefaultSessionConfig liefert die Standard-Session-Einstellungen zurück.
func DefaultSessionConfig() neo4j.SessionConfig {
	return neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
	}
}
