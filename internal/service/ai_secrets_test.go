package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dylanbr0wn/shiet/internal/integration/secrets"
	"github.com/dylanbr0wn/shiet/internal/service"
)

type memoryTokenStore struct {
	tokens map[string]secrets.Token
}

func (m *memoryTokenStore) key(provider, accountID string) string {
	return provider + ":" + accountID
}

func (m *memoryTokenStore) Get(provider, accountID string) (secrets.Token, error) {
	token, ok := m.tokens[m.key(provider, accountID)]
	if !ok {
		return secrets.Token{}, secrets.ErrNotFound
	}
	return token, nil
}

func (m *memoryTokenStore) Set(provider, accountID string, token secrets.Token) error {
	if m.tokens == nil {
		m.tokens = make(map[string]secrets.Token)
	}
	m.tokens[m.key(provider, accountID)] = token
	return nil
}

func (m *memoryTokenStore) Delete(provider, accountID string) error {
	delete(m.tokens, m.key(provider, accountID))
	return nil
}

func TestAIAPIKeyKeychainCRUD(t *testing.T) {
	e := newSyncEnv(t)
	store := &memoryTokenStore{tokens: make(map[string]secrets.Token)}
	e.svc.SetAISecrets(store)
	ctx := context.Background()

	if e.svc.HasAIAPIKey(ctx) {
		t.Fatal("expected no key initially")
	}
	if err := e.svc.SaveAIAPIKey(ctx, "sk-test"); err != nil {
		t.Fatalf("SaveAIAPIKey: %v", err)
	}
	if !e.svc.HasAIAPIKey(ctx) {
		t.Fatal("expected stored key")
	}
	if err := e.svc.ClearAIAPIKey(ctx); err != nil {
		t.Fatalf("ClearAIAPIKey: %v", err)
	}
	if e.svc.HasAIAPIKey(ctx) {
		t.Fatal("expected key cleared")
	}
}

func TestSync_UsesStoredAIAPIKey(t *testing.T) {
	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "Meetings"}},
			},
		})
	}))
	defer server.Close()

	e := newSyncEnv(t)
	store := &memoryTokenStore{tokens: make(map[string]secrets.Token)}
	e.svc.SetAISecrets(store)
	ctx := context.Background()

	if err := e.svc.SaveAIAPIKey(ctx, "sk-test"); err != nil {
		t.Fatalf("SaveAIAPIKey: %v", err)
	}
	if err := e.svc.SaveAIConfig(ctx, server.URL+"/v1", "llama3"); err != nil {
		t.Fatalf("SaveAIConfig: %v", err)
	}

	event := e.baseEvent()
	event.Title = "Quarterly planning"
	if _, err := e.svc.SyncEvents(ctx, e.periodID, []service.IncomingEvent{event}); err != nil {
		t.Fatalf("SyncEvents: %v", err)
	}
	if authHeader != "Bearer sk-test" {
		t.Fatalf("Authorization = %q want Bearer sk-test", authHeader)
	}
}
