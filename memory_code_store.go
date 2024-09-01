package emailauth

import (
	"context"
	"sync"
	"time"
)

type InMemoryCodeStore struct {
	mu    sync.RWMutex
	codes map[string]*AuthCode
}

// NewInMemoryCodeStore initializes and returns a new InMemoryCodeStore.
func NewInMemoryCodeStore() *InMemoryCodeStore {
	return &InMemoryCodeStore{
		codes: make(map[string]*AuthCode),
	}
}

// Set stores the authentication code for a given email in memory.
func (s *InMemoryCodeStore) Set(ctx context.Context, email string, code *AuthCode) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.codes[email] = code

	// Launch a goroutine to delete the code after it expires
	go s.expireCode(email, code.ExpiresAt)

	return nil
}

// Get retrieves the authentication code for a given email from memory.
func (s *InMemoryCodeStore) Get(ctx context.Context, email string) (*AuthCode, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	code, exists := s.codes[email]
	if !exists {
		return nil, nil
	}

	return code, nil
}

// Delete removes the authentication code for a given email from memory.
func (s *InMemoryCodeStore) Delete(ctx context.Context, email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.codes, email)
	return nil
}

// expireCode deletes the code from memory after its expiration time.
func (s *InMemoryCodeStore) expireCode(email string, expiresAt time.Time) {
	timer := time.NewTimer(time.Until(expiresAt))
	defer timer.Stop()

	<-timer.C

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check the code still exists and hasn't been updated
	if code, exists := s.codes[email]; exists && time.Now().After(code.ExpiresAt) {
		delete(s.codes, email)
	}
}
