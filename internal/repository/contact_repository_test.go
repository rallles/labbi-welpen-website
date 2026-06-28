package repository

import (
	"context"
	"errors"
	"testing"
)

type stubRecordCursor struct {
	found bool
	err   error
}

func (s stubRecordCursor) Next(context.Context) bool { return s.found }
func (s stubRecordCursor) Err() error                { return s.err }

func TestRequireContactUpdateResult(t *testing.T) {
	t.Run("contact found", func(t *testing.T) {
		if err := requireContactUpdateResult(context.Background(), stubRecordCursor{found: true}, "contact-1"); err != nil {
			t.Fatalf("requireContactUpdateResult() error = %v", err)
		}
	})

	t.Run("contact missing", func(t *testing.T) {
		err := requireContactUpdateResult(context.Background(), stubRecordCursor{}, "missing")
		if !errors.Is(err, ErrContactNotFound) {
			t.Fatalf("error = %v, want ErrContactNotFound", err)
		}
	})

	t.Run("cursor error", func(t *testing.T) {
		cursorErr := errors.New("cursor failed")
		err := requireContactUpdateResult(context.Background(), stubRecordCursor{err: cursorErr}, "contact-1")
		if !errors.Is(err, cursorErr) {
			t.Fatalf("error = %v, want wrapped cursor error", err)
		}
	})
}
