package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const anthropicVersion = "2023-06-01"

// AnthropicClient talks to Anthropic's Messages API.
type AnthropicClient struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

// NewAnthropicClient builds a client for Anthropic's HTTP API.
func NewAnthropicClient(baseURL, apiKey string) *AnthropicClient {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "https://api.anthropic.com"
	}
	return &AnthropicClient{
		BaseURL: base,
		APIKey:  strings.TrimSpace(apiKey),
		HTTP:    &http.Client{Timeout: defaultTimeout},
	}
}

type anthropicModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// ListModels fetches model ids from GET /v1/models.
func (c *AnthropicClient) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/v1/models", nil)
	if err != nil {
		return nil, err
	}
	c.applyHeaders(req)

	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return nil, fmt.Errorf("list models: %s: %s", res.Status, strings.TrimSpace(string(body)))
	}

	var payload anthropicModelsResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}

	out := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		if item.ID != "" {
			out = append(out, item.ID)
		}
	}
	return out, nil
}

type anthropicMessageRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicMessageResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// Validate sends a tiny message to confirm the endpoint and model work.
func (c *AnthropicClient) Validate(ctx context.Context, model string) error {
	_, err := c.ChatCompletion(ctx, model, "", "Hi", 32)
	return err
}

// ChatCompletion sends a message request and returns the assistant text.
func (c *AnthropicClient) ChatCompletion(ctx context.Context, model, systemPrompt, userPrompt string, maxTokens int) (string, error) {
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}

	body, err := json.Marshal(anthropicMessageRequest{
		Model:     model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: userPrompt},
		},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	c.applyHeaders(req)

	res, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return "", fmt.Errorf("chat completion: %s: %s", res.Status, strings.TrimSpace(string(payload)))
	}

	var payload anthropicMessageResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return "", err
	}
	for _, block := range payload.Content {
		if strings.TrimSpace(block.Text) != "" {
			return strings.TrimSpace(block.Text), nil
		}
	}
	return "", fmt.Errorf("chat completion: empty response")
}

func (c *AnthropicClient) applyHeaders(req *http.Request) {
	if c.APIKey != "" {
		req.Header.Set("x-api-key", c.APIKey)
	}
	req.Header.Set("anthropic-version", anthropicVersion)
}
