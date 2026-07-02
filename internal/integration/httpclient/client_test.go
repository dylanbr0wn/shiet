package httpclient_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dylanbr0wn/clockr/internal/integration/httpclient"
	"github.com/dylanbr0wn/clockr/internal/integration/oauth"
	"github.com/dylanbr0wn/clockr/internal/integration/secrets"
)

func TestClientInjectsBearerToken(t *testing.T) {
	store := secrets.NewMemoryStore()
	_ = store.Set("github", "octocat", secrets.Token{
		AccessToken: "secret-token",
		TokenType:   "Bearer",
	})

	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.Client{
		Provider:  "github",
		AccountID: "octocat",
		Store:     store,
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	resp, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if authHeader != "Bearer secret-token" {
		t.Fatalf("authorization: %q", authHeader)
	}
}

func TestClientRefreshesOn401(t *testing.T) {
	store := secrets.NewMemoryStore()
	_ = store.Set("google", "user@example.com", secrets.Token{
		AccessToken:  "expired",
		RefreshToken: "refresh",
		Expiry:       time.Now().Add(-time.Hour),
	})

	var calls atomic.Int32
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"access_token":"fresh","token_type":"Bearer","expires_in":3600}`)
	}))
	defer tokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Authorization") != "Bearer fresh" {
			t.Fatalf("expected refreshed token, got %q", r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	client := httpclient.Client{
		Provider:  "google",
		AccountID: "user@example.com",
		Config: oauth.ProviderConfig{
			Provider: "google",
			ClientID: "client-id",
			TokenURL: tokenServer.URL + "/token",
		},
		Store: store,
	}

	req, _ := http.NewRequest(http.MethodGet, apiServer.URL, nil)
	resp, err := client.Do(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: %d", resp.StatusCode)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls: %d", calls.Load())
	}

	got, err := store.Get("google", "user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "fresh" {
		t.Fatalf("stored token: %+v", got)
	}
}

func TestClientRetriesRateLimit(t *testing.T) {
	store := secrets.NewMemoryStore()
	_ = store.Set("slack", "workspace", secrets.Token{AccessToken: "token"})

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	}))
	defer server.Close()

	client := httpclient.Client{
		Provider:  "slack",
		AccountID: "workspace",
		Store:     store,
	}

	req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.Do(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := httpclient.ReadBody(resp)
	if string(body) != "ok" {
		t.Fatalf("body: %q", body)
	}
	if calls.Load() != 2 {
		t.Fatalf("calls: %d", calls.Load())
	}
}
