package security

import "testing"

func TestCSRFGenerateProducesValidToken(t *testing.T) {
	store := NewCSRFStore()

	token, err := store.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if token == "" {
		t.Fatal("Generate() token is empty")
	}
	if !store.Valid(token) {
		t.Fatal("Valid() = false, want true for generated token")
	}
}

func TestCSRFValidRejectsEmptyToken(t *testing.T) {
	store := NewCSRFStore()
	if store.Valid("") {
		t.Fatal("Valid(\"\") = true, want false")
	}
}

func TestCSRFValidRejectsUnknownToken(t *testing.T) {
	store := NewCSRFStore()
	if store.Valid("unknown-token") {
		t.Fatal("Valid(unknown-token) = true, want false")
	}
}

type csrfStoreWithConsume interface {
	Consume(token string) bool
}

func TestCSRFConsumeInvalidatesTokenWhenSupported(t *testing.T) {
	store := NewCSRFStore()
	token, err := store.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	consumer, ok := any(store).(csrfStoreWithConsume)
	if !ok {
		t.Skip("Consume() is not implemented")
	}
	if !consumer.Consume(token) {
		t.Fatal("Consume() = false, want true for generated token")
	}
	if store.Valid(token) {
		t.Fatal("Valid() = true after Consume(), want false")
	}
}