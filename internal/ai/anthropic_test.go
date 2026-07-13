package ai_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dylanbr0wn/shiet/internal/ai"
)

func TestNewChatClientRoutesAnthropic(t *testing.T) {
	t.Parallel()
	if _, ok := ai.NewChatClient("https://api.anthropic.com", "key").(*ai.AnthropicClient); !ok {
		t.Fatal("expected AnthropicClient")
	}
	if _, ok := ai.NewChatClient("http://127.0.0.1:1234/v1", "").(*ai.Client); !ok {
		t.Fatal("expected OpenAI-compatible Client")
	}
}

func TestAnthropicChatCompletion(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/messages" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("x-api-key"); got != "secret" {
			t.Fatalf("x-api-key = %q want secret", got)
		}
		if got := r.Header.Get("anthropic-version"); got == "" {
			t.Fatal("missing anthropic-version header")
		}

		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Model     string `json:"model"`
			MaxTokens int    `json:"max_tokens"`
			System    string `json:"system"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Model != "claude-test" || payload.MaxTokens != 128 {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		if payload.System != "sys" {
			t.Fatalf("system = %q want sys", payload.System)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]string{
				{"type": "text", "text": "Meetings"},
			},
		})
	}))
	defer server.Close()

	client := ai.NewAnthropicClient(server.URL, "secret")
	reply, err := client.ChatCompletion(context.Background(), "claude-test", "sys", "user", 128)
	if err != nil {
		t.Fatalf("ChatCompletion: %v", err)
	}
	if reply != "Meetings" {
		t.Fatalf("reply = %q", reply)
	}
}

func TestAnthropicListModels(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]string{{"id": "claude-3-5-sonnet-20240620"}},
		})
	}))
	defer server.Close()

	client := ai.NewAnthropicClient(server.URL, "secret")
	models, err := client.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if len(models) != 1 || models[0] != "claude-3-5-sonnet-20240620" {
		t.Fatalf("models = %#v", models)
	}
}

func TestIsAnthropicEndpoint(t *testing.T) {
	t.Parallel()
	cases := map[string]bool{
		"https://api.anthropic.com":    true,
		"https://api.anthropic.com/v1": true,
		"https://api.openai.com/v1":    false,
		"http://127.0.0.1:1234/v1":     false,
	}
	for raw, want := range cases {
		if got := ai.IsAnthropicEndpoint(raw); got != want {
			t.Fatalf("IsAnthropicEndpoint(%q) = %v want %v", raw, got, want)
		}
	}
}

func TestAnthropicValidateRequiresReply(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content": []map[string]string{
				{"type": "text", "text": " "},
			},
		})
	}))
	defer server.Close()

	client := ai.NewAnthropicClient(server.URL, "secret")
	err := client.Validate(context.Background(), "claude-test")
	if err == nil || !strings.Contains(err.Error(), "empty response") {
		t.Fatalf("Validate err = %v", err)
	}
}
