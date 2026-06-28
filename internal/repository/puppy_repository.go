package repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"labbi-app/internal/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
)

var ErrPuppyNotFound = errors.New("puppy not found")

type PuppyRepository struct {
	driver neo4j.DriverWithContext
}

func NewPuppyRepository(driver neo4j.DriverWithContext) *PuppyRepository {
	return &PuppyRepository{driver: driver}
}

func (r *PuppyRepository) Create(ctx context.Context, puppy models.Puppy) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			CREATE (p:Puppy {
				id: $id,
				name: $name,
				geburtsdatum: date($geburtsdatum),
				geschlecht: $geschlecht,
				farbe: $farbe,
				gewicht: $gewicht,
				charakter: $charakter,
				geimpft: $geimpft,
				gechippt: $gechippt,
				entwurmt: $entwurmt,
				eltern: $eltern,
				notizen: $notizen,
				bilder: $bilder
			})`, puppyParams(puppy))
		if err != nil {
			return nil, err
		}
		if _, err := result.Consume(ctx); err != nil {
			return nil, err
		}
		if err := replaceParentRelationships(ctx, tx, puppy.ID, puppy.Eltern); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func (r *PuppyRepository) List(ctx context.Context) ([]models.Puppy, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx, `
		MATCH (p:Puppy)
		RETURN p
		ORDER BY p.geburtsdatum DESC, p.name ASC`, nil)
	if err != nil {
		return nil, err
	}

	puppies := make([]models.Puppy, 0)
	for result.Next(ctx) {
		puppy, err := puppyFromRecord(result.Record())
		if err != nil {
			return nil, err
		}
		puppies = append(puppies, puppy)
	}
	if err := result.Err(); err != nil {
		return nil, err
	}
	return puppies, nil
}

func (r *PuppyRepository) GetByID(ctx context.Context, id string) (models.Puppy, error) {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer session.Close(ctx)

	result, err := session.Run(ctx, `MATCH (p:Puppy {id: $id}) RETURN p`, map[string]any{"id": id})
	if err != nil {
		return models.Puppy{}, err
	}
	if !result.Next(ctx) {
		if err := result.Err(); err != nil {
			return models.Puppy{}, err
		}
		return models.Puppy{}, ErrPuppyNotFound
	}
	return puppyFromRecord(result.Record())
}

func (r *PuppyRepository) Update(ctx context.Context, puppy models.Puppy) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (p:Puppy {id: $id})
			SET p.name = $name,
				p.geburtsdatum = date($geburtsdatum),
				p.geschlecht = $geschlecht,
				p.farbe = $farbe,
				p.gewicht = $gewicht,
				p.charakter = $charakter,
				p.geimpft = $geimpft,
				p.gechippt = $gechippt,
				p.entwurmt = $entwurmt,
				p.eltern = $eltern,
				p.notizen = $notizen,
				p.bilder = $bilder
			RETURN p.id`, puppyParams(puppy))
		if err != nil {
			return nil, err
		}
		if !result.Next(ctx) {
			if err := result.Err(); err != nil {
				return nil, err
			}
			return nil, ErrPuppyNotFound
		}
		if err := replaceParentRelationships(ctx, tx, puppy.ID, puppy.Eltern); err != nil {
			return nil, err
		}
		return nil, result.Err()
	})
	return err
}

func (r *PuppyRepository) Delete(ctx context.Context, id string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (p:Puppy {id: $id})
			WITH p, p.id AS deletedID
			DETACH DELETE p
			RETURN deletedID`, map[string]any{"id": id})
		if err != nil {
			return nil, err
		}
		if err := requirePuppyDeleteResult(ctx, result, id); err != nil {
			return nil, err
		}
		if _, err := result.Consume(ctx); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

type puppyDeleteCursor interface {
	Next(context.Context) bool
	Err() error
}

func requirePuppyDeleteResult(ctx context.Context, result puppyDeleteCursor, id string) error {
	if result.Next(ctx) {
		return nil
	}
	if err := result.Err(); err != nil {
		return fmt.Errorf("read puppy %q delete result: %w", id, err)
	}
	return fmt.Errorf("%w: %s", ErrPuppyNotFound, id)
}

func replaceParentRelationships(ctx context.Context, tx neo4j.ManagedTransaction, puppyID string, parents []string) error {
	result, err := tx.Run(ctx, `MATCH (:Puppy {id: $id})-[r:HAS_PARENT]->(:Dog) DELETE r`, map[string]any{"id": puppyID})
	if err != nil {
		return err
	}
	if _, err := result.Consume(ctx); err != nil {
		return err
	}

	for _, rawParentID := range parents {
		parentID := models.NormalizeParentDogID(strings.TrimSpace(rawParentID))
		if parentID == "" {
			continue
		}
		result, err := tx.Run(ctx, `
			MATCH (p:Puppy {id: $puppyID})
			MATCH (parent:Dog {id: $parentID})
			MERGE (p)-[:HAS_PARENT]->(parent)
			RETURN parent.id`, map[string]any{
			"puppyID":  puppyID,
			"parentID": parentID,
		})
		if err != nil {
			return err
		}
		if !result.Next(ctx) {
			if err := result.Err(); err != nil {
				return err
			}
			return fmt.Errorf("parent dog %q not found", parentID)
		}
		if err := result.Err(); err != nil {
			return err
		}
	}
	return nil
}

func puppyParams(p models.Puppy) map[string]any {
	return map[string]any{
		"id":           p.ID,
		"name":         p.Name,
		"geburtsdatum": p.Geburtsdatum,
		"geschlecht":   p.Geschlecht,
		"farbe":        string(p.Farbe),
		"gewicht":      p.Gewicht,
		"charakter":    p.Charakter,
		"geimpft":      p.Geimpft,
		"gechippt":     p.Gechippt,
		"entwurmt":     p.Entwurmt,
		"eltern":       normalizeParentDogIDs(p.Eltern),
		"notizen":      p.Notizen,
		"bilder":       p.Bilder,
	}
}

func puppyFromRecord(record *neo4j.Record) (models.Puppy, error) {
	value, ok := record.Get("p")
	if !ok {
		return models.Puppy{}, fmt.Errorf("puppy node missing in result")
	}

	node, ok := value.(dbtype.Node)
	if !ok {
		return models.Puppy{}, fmt.Errorf("unexpected puppy node type %T", value)
	}

	props := node.Props
	return models.Puppy{
		ID:           stringValue(props["id"]),
		Name:         stringValue(props["name"]),
		Geburtsdatum: dateString(props["geburtsdatum"]),
		Geschlecht:   stringValue(props["geschlecht"]),
		Farbe:        models.Fellfarbe(stringValue(props["farbe"])),
		Gewicht:      floatValue(props["gewicht"]),
		Charakter:    stringValue(props["charakter"]),
		Geimpft:      boolValue(props["geimpft"]),
		Gechippt:     boolValue(props["gechippt"]),
		Entwurmt:     boolValue(props["entwurmt"]),
		Eltern:       normalizeParentDogIDs(stringSlice(props["eltern"])),
		Notizen:      stringValue(props["notizen"]),
		Bilder:       stringSlice(props["bilder"]),
	}, nil
}

func normalizeParentDogIDs(values []string) []string {
	parents := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		id := models.NormalizeParentDogID(strings.TrimSpace(value))
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		parents = append(parents, id)
	}
	return parents
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", value)
}

func dateString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case dbtype.Date:
		return v.Time().Format("2006-01-02")
	case time.Time:
		return v.Format("2006-01-02")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func floatValue(value any) float64 {
	switch v := value.(type) {
	case nil:
		return 0
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0
	}
}

func boolValue(value any) bool {
	switch v := value.(type) {
	case nil:
		return false
	case bool:
		return v
	case string:
		return v == "true" || v == "on" || v == "1"
	default:
		return false
	}
}

func stringSlice(value any) []string {
	switch v := value.(type) {
	case nil:
		return nil
	case []string:
		return v
	case []any:
		items := make([]string, 0, len(v))
		for _, item := range v {
			items = append(items, stringValue(item))
		}
		return items
	default:
		return []string{stringValue(v)}
	}
}
