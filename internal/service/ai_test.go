package service_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dylanbr0wn/clockr/internal/db/sqlc"
	"github.com/dylanbr0wn/clockr/internal/service"
)

func TestSync_AppliesAISuggestion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "Meetings"}},
			},
		})
	}))
	defer server.Close()

	e := newSyncEnv(t)
	ctx := context.Background()

	if err := e.svc.SaveAIConfig(ctx, server.URL+"/v1", "llama3"); err != nil {
		t.Fatalf("SaveAIConfig: %v", err)
	}

	event := e.baseEvent()
	event.Title = "Quarterly planning"
	res, err := e.svc.SyncEvents(ctx, e.periodID, []service.IncomingEvent{event})
	if err != nil {
		t.Fatalf("SyncEvents: %v", err)
	}
	if res.Added != 1 {
		t.Fatalf("expected 1 added, got %+v", res)
	}

	cats, err := e.svc.ListCategories(ctx)
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	meetingsID := e.catID
	for _, cat := range cats {
		if cat.Name == "Meetings" {
			meetingsID = cat.ID
			break
		}
	}

	o, err := e.q.GetOverlay(ctx, sqlc.GetOverlayParams{
		PeriodID:      e.periodID,
		GoogleEventID: event.GoogleEventID,
		InstanceID:    "",
		Kind:          "category",
	})
	if err != nil {
		t.Fatalf("expected ai overlay: %v", err)
	}
	if !o.CategoryID.Valid || o.CategoryID.Int64 != meetingsID {
		t.Fatalf("overlay category mismatch: %+v want %d", o, meetingsID)
	}
}

func TestValidateAIConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "ok"}},
			},
		})
	}))
	defer server.Close()

	e := newSyncEnv(t)
	ctx := context.Background()

	result, err := e.svc.ValidateAIConfig(ctx, server.URL+"/v1", "", "llama3")
	if err != nil {
		t.Fatalf("ValidateAIConfig: %v", err)
	}
	if !result.OK {
		t.Fatalf("expected ok result, got %+v", result)
	}
}
