package ai

import (
	"context"
	"net/url"
	"strings"
)

// ChatClient talks to a remote or local model for listing, validation, and chat.
type ChatClient interface {
	ListModels(ctx context.Context) ([]string, error)
	Validate(ctx context.Context, model string) error
	ChatCompletion(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int) (string, error)
}

// NewChatClient picks an Anthropic or OpenAI-compatible client for the base URL.
func NewChatClient(baseURL, apiKey string) ChatClient {
	if IsAnthropicEndpoint(baseURL) {
		return NewAnthropicClient(baseURL, apiKey)
	}
	return NewClient(baseURL, apiKey)
}

// IsAnthropicEndpoint reports whether baseURL targets Anthropic's API.
func IsAnthropicEndpoint(baseURL string) bool {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return false
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return strings.Contains(strings.ToLower(baseURL), "anthropic.com")
	}
	host := strings.ToLower(parsed.Hostname())
	return host == "api.anthropic.com" || strings.HasSuffix(host, ".anthropic.com")
}
