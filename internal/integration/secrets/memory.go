package secrets

import (
	"sync"
)

// MemoryStore is an in-memory TokenStore for tests and local development.
type MemoryStore struct {
	mu     sync.RWMutex
	tokens map[string]Token
}

// NewMemoryStore returns an empty in-memory token store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tokens: make(map[string]Token)}
}

// Get retrieves a token from memory.
func (s *MemoryStore) Get(provider, accountID string) (Token, error) {
	key, err := storageKey(provider, accountID)
	if err != nil {
		return Token{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	token, ok := s.tokens[key]
	if !ok {
		return Token{}, ErrNotFound
	}
	return token, nil
}

// Set stores a token in memory.
func (s *MemoryStore) Set(provider, accountID string, token Token) error {
	key, err := storageKey(provider, accountID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[key] = token
	return nil
}

// Delete removes a token from memory.
func (s *MemoryStore) Delete(provider, accountID string) error {
	key, err := storageKey(provider, accountID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, key)
	return nil
}
