package security

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

const csrfTokenTTL = 2 * time.Hour

type CSRFStore struct {
	mu     sync.Mutex
	tokens map[string]time.Time
}

func NewCSRFStore() *CSRFStore {
	return &CSRFStore{tokens: make(map[string]time.Time)}
}

func (s *CSRFStore) Generate() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cleanupLocked(time.Now())
	s.tokens[token] = time.Now().Add(csrfTokenTTL)
	return token, nil
}

func (s *CSRFStore) Valid(token string) bool {
	if token == "" {
		return false
	}

	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	expiresAt, ok := s.tokens[token]
	if !ok {
		return false
	}
	if now.After(expiresAt) {
		delete(s.tokens, token)
		return false
	}
	return true
}

func (s *CSRFStore) cleanupLocked(now time.Time) {
	for token, expiresAt := range s.tokens {
		if now.After(expiresAt) {
			delete(s.tokens, token)
		}
	}
}

var DefaultCSRF = NewCSRFStore()
