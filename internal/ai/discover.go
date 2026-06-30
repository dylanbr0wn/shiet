package ai

import (
	"context"
	"sync"
	"time"
)

var knownLocalEndpoints = []struct {
	name    string
	baseURL string
}{
	{name: "Ollama", baseURL: "http://127.0.0.1:11434/v1"},
	{name: "LM Studio", baseURL: "http://127.0.0.1:1234/v1"},
}

// DiscoverLocalEndpoints probes known local OpenAI-compatible ports and returns
// whichever endpoints are currently reachable.
func DiscoverLocalEndpoints(ctx context.Context) []Endpoint {
	results := make([]Endpoint, len(knownLocalEndpoints))
	var wg sync.WaitGroup

	for i, candidate := range knownLocalEndpoints {
		wg.Add(1)
		go func(i int, candidate struct {
			name    string
			baseURL string
		}) {
			defer wg.Done()
			local, verdict := ClassifyEndpoint(candidate.baseURL)
			endpoint := Endpoint{
				Name:    candidate.name,
				BaseURL: candidate.baseURL,
				Local:   local,
			}
			_ = verdict

			probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			client := NewClient(candidate.baseURL, "")
			models, err := client.ListModels(probeCtx)
			if err == nil {
				endpoint.Running = true
				endpoint.Models = models
			}
			results[i] = endpoint
		}(i, candidate)
	}

	wg.Wait()
	return results
}
