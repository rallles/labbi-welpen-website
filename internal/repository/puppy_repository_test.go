package repository

import (
	"context"
	"errors"
	"testing"
)

type stubPuppyDeleteCursor struct {
	found bool
	err   error
}

func (s stubPuppyDeleteCursor) Next(context.Context) bool { return s.found }
func (s stubPuppyDeleteCursor) Err() error                { return s.err }

func TestRequirePuppyDeleteResult(t *testing.T) {
	t.Run("puppy deleted", func(t *testing.T) {
		if err := requirePuppyDeleteResult(context.Background(), stubPuppyDeleteCursor{found: true}, "puppy-1"); err != nil {
			t.Fatalf("requirePuppyDeleteResult() error = %v", err)
		}
	})

	t.Run("puppy missing", func(t *testing.T) {
		err := requirePuppyDeleteResult(context.Background(), stubPuppyDeleteCursor{}, "missing")
		if !errors.Is(err, ErrPuppyNotFound) {
			t.Fatalf("error = %v, want ErrPuppyNotFound", err)
		}
	})

	t.Run("cursor error", func(t *testing.T) {
		cursorErr := errors.New("cursor failed")
		err := requirePuppyDeleteResult(context.Background(), stubPuppyDeleteCursor{err: cursorErr}, "puppy-1")
		if !errors.Is(err, cursorErr) {
			t.Fatalf("error = %v, want wrapped cursor error", err)
		}
	})
}
