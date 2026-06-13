package repository

import (
	"context"

	"labbi-app/internal/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type ContactRepository struct {
	driver neo4j.DriverWithContext
}

func NewContactRepository(driver neo4j.DriverWithContext) *ContactRepository {
	return &ContactRepository{driver: driver}
}

func (r *ContactRepository) Create(ctx context.Context, contact models.Contact) error {
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)

	result, err := session.Run(ctx, `
		CREATE (c:Contact {
			id: $id,
			name: $name,
			email: $email,
			phone: $phone,
			message: $message,
			createdAt: datetime($createdAt),
			mailSent: $mailSent,
			mailError: $mailError
		})`, map[string]any{
		"id":        contact.ID,
		"name":      contact.Name,
		"email":     contact.Email,
		"phone":     contact.Phone,
		"message":   contact.Message,
		"createdAt": contact.CreatedAt.Format(timeFormatRFC3339Nano),
		"mailSent":  contact.MailSent,
		"mailError": contact.MailError,
	})
	if err != nil {
		return err
	}
	_, err = result.Consume(ctx)
	return err
}

const timeFormatRFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
