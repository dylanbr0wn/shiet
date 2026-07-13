package bitbucket_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dylanbr0wn/shiet/internal/db"
	"github.com/dylanbr0wn/shiet/internal/db/sqlc"
	"github.com/dylanbr0wn/shiet/internal/integration/bitbucket"
	"github.com/dylanbr0wn/shiet/internal/integration/connection"
	"github.com/dylanbr0wn/shiet/internal/integration/secrets"
)

func TestSyncWorkspacesReposPreservesSelection(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/user"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"uuid":         "{user-1}",
				"display_name": "Dylan",
				"username":     "dylan",
			})
		case strings.Contains(r.URL.Path, "/workspaces"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"values": []map[string]any{{
					"uuid": "{ws-1}",
					"slug": "acme",
					"name": "Acme",
				}},
			})
		case strings.Contains(r.URL.Path, "/repositories"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"values": []map[string]any{{
					"uuid":       "{repo-1}",
					"name":       "app",
					"full_name":  "acme/app",
					"is_private": true,
					"workspace": map[string]any{
						"uuid": "{ws-1}",
						"slug": "acme",
					},
				}},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := db.Migrate(conn); err != nil {
		t.Fatal(err)
	}

	registry := connection.NewRegistry(conn)
	store := secrets.NewMemoryStore()
	queries := sqlc.New(conn)
	provider := &bitbucket.Provider{
		Config:   bitbucket.OAuthConfig("client", "secret"),
		Store:    store,
		Registry: registry,
		Queries:  queries,
		HTTP:     srv.Client(),
		BaseURL:  srv.URL,
		AuthMode: "local",
	}

	if err := store.Set("bitbucket", "{user-1}", secrets.Token{AccessToken: "token", TokenType: "Bearer"}); err != nil {
		t.Fatal(err)
	}
	if _, err := registry.Upsert(context.Background(), connection.UpsertInput{
		Provider:     "bitbucket",
		AccountLabel: "Dylan",
		AccountID:    "{user-1}",
		Status:       connection.StatusConnected,
	}); err != nil {
		t.Fatal(err)
	}

	workspaces, repos, err := provider.SyncWorkspacesRepos(context.Background(), "{user-1}")
	if err != nil {
		t.Fatal(err)
	}
	if len(workspaces) != 1 || len(repos) != 1 {
		t.Fatalf("sync = %d workspaces, %d repos", len(workspaces), len(repos))
	}

	if err := provider.SetWorkspaceSelected(context.Background(), workspaces[0].ID, true); err != nil {
		t.Fatal(err)
	}
	if err := provider.SetRepoSelected(context.Background(), repos[0].ID, true); err != nil {
		t.Fatal(err)
	}

	workspaces, repos, err = provider.SyncWorkspacesRepos(context.Background(), "{user-1}")
	if err != nil {
		t.Fatal(err)
	}
	if workspaces[0].Selected != 1 || repos[0].Selected != 1 {
		t.Fatalf("selection not preserved: workspace=%d repo=%d", workspaces[0].Selected, repos[0].Selected)
	}
}
