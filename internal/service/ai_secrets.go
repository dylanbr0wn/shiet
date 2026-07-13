package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/dylanbr0wn/shiet/internal/ai"
	"github.com/dylanbr0wn/shiet/internal/integration/secrets"
)

const (
	aiSecretProvider  = "ai"
	aiSecretAccountID = "default"
)

// SetAISecrets wires the OS keychain store used for AI API keys.
func (s *Service) SetAISecrets(store secrets.TokenStore) {
	if s == nil {
		return
	}
	s.aiSecrets = store
}

// SaveAIAPIKey stores an AI API key in the OS keychain.
func (s *Service) SaveAIAPIKey(_ context.Context, apiKey string) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return fmt.Errorf("save ai api key: key is required")
	}
	if s.aiSecrets == nil {
		return fmt.Errorf("save ai api key: keychain unavailable")
	}
	return s.aiSecrets.Set(aiSecretProvider, aiSecretAccountID, secrets.Token{
		AccessToken: apiKey,
		TokenType:   "Bearer",
	})
}

// ClearAIAPIKey removes the stored AI API key from the OS keychain.
func (s *Service) ClearAIAPIKey(_ context.Context) error {
	if s.aiSecrets == nil {
		return nil
	}
	return s.aiSecrets.Delete(aiSecretProvider, aiSecretAccountID)
}

// HasAIAPIKey reports whether an AI API key is stored in the keychain.
func (s *Service) HasAIAPIKey(_ context.Context) bool {
	return s.loadAIAPIKey() != ""
}

func (s *Service) loadAIAPIKey() string {
	if s == nil || s.aiSecrets == nil {
		return ""
	}
	token, err := s.aiSecrets.Get(aiSecretProvider, aiSecretAccountID)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(token.AccessToken)
}

func (s *Service) resolveAIAPIKey(provided string) string {
	if key := strings.TrimSpace(provided); key != "" {
		return key
	}
	return s.loadAIAPIKey()
}

func (s *Service) loadAIChatClient(ctx context.Context, baseURL, providedKey string) ai.ChatClient {
	_ = ctx
	return ai.NewChatClient(baseURL, s.resolveAIAPIKey(providedKey))
}
