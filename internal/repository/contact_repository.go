package repository

import (
	"context"
	"errors"
	"fmt"

	"labbi-app/internal/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var ErrContactNotFound = errors.New("contact not found")

type ContactRepository struct {
	driver neo4j.DriverWithContext
}

func NewContactRepository(driver neo4j.DriverWithContext) *ContactRepository {
	return &ContactRepository{driver: driver}
}

func (r *ContactRepository) Create(ctx context.Context, contact models.Contact) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			CREATE (c:Contact {
				id: $id,
				name: $name,
				email: $email,
				phone: $phone,
				message: $message,
				createdAt: datetime($createdAt),
				mailSent: $mailSent,
				mailError: $mailError
			})`, contactParams(contact))
		if err != nil {
			return nil, err
		}
		_, err = result.Consume(ctx)
		return nil, err
	})
	return err
}

func (r *ContactRepository) UpdateMailStatus(ctx context.Context, id string, sent bool, mailError string) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
			MATCH (c:Contact {id: $id})
			SET c.mailSent = $mailSent,
				c.mailError = $mailError
			RETURN c.id`,
			map[string]any{
				"id":        id,
				"mailSent":  sent,
				"mailError": mailError,
			})
		if err != nil {
			return nil, fmt.Errorf("update contact %q mail status: %w", id, err)
		}
		if err := requireContactUpdateResult(ctx, result, id); err != nil {
			return nil, err
		}
		if _, err := result.Consume(ctx); err != nil {
			return nil, fmt.Errorf("consume contact %q mail status update: %w", id, err)
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("update mail status for contact %q: %w", id, err)
	}
	return nil
}

type recordCursor interface {
	Next(context.Context) bool
	Err() error
}

func requireContactUpdateResult(ctx context.Context, result recordCursor, id string) error {
	if result.Next(ctx) {
		return nil
	}
	if err := result.Err(); err != nil {
		return fmt.Errorf("read contact %q mail status update result: %w", id, err)
	}
	return fmt.Errorf("%w: %s", ErrContactNotFound, id)
}

func contactParams(contact models.Contact) map[string]any {
	return map[string]any{
		"id":        contact.ID,
		"name":      contact.Name,
		"email":     contact.Email,
		"phone":     contact.Phone,
		"message":   contact.Message,
		"createdAt": contact.CreatedAt.Format(timeFormatRFC3339Nano),
		"mailSent":  contact.MailSent,
		"mailError": contact.MailError,
	}
}

const timeFormatRFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
