package secrets

import (
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

// KeyringStore persists tokens in the OS keychain via go-keyring.
type KeyringStore struct {
	Service string
}

// NewKeyringStore returns a store backed by the OS keychain.
func NewKeyringStore() *KeyringStore {
	return &KeyringStore{Service: keyringService}
}

func (s *KeyringStore) serviceName() string {
	if s != nil && s.Service != "" {
		return s.Service
	}
	return keyringService
}

// Get retrieves a token from the keychain.
func (s *KeyringStore) Get(provider, accountID string) (Token, error) {
	key, err := storageKey(provider, accountID)
	if err != nil {
		return Token{}, err
	}

	raw, err := keyring.Get(s.serviceName(), key)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return Token{}, ErrNotFound
		}
		return Token{}, fmt.Errorf("keyring get: %w", err)
	}
	return decodeToken(raw)
}

// Set stores a token in the keychain.
func (s *KeyringStore) Set(provider, accountID string, token Token) error {
	key, err := storageKey(provider, accountID)
	if err != nil {
		return err
	}

	raw, err := encodeToken(token)
	if err != nil {
		return err
	}
	if err := keyring.Set(s.serviceName(), key, raw); err != nil {
		return fmt.Errorf("keyring set: %w", err)
	}
	return nil
}

// Delete removes a token from the keychain.
func (s *KeyringStore) Delete(provider, accountID string) error {
	key, err := storageKey(provider, accountID)
	if err != nil {
		return err
	}
	if err := keyring.Delete(s.serviceName(), key); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("keyring delete: %w", err)
	}
	return nil
}
