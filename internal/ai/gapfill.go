package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// SuggestGapFill asks the model to pick a category and write a short timesheet
// description for an uncovered interval, optionally citing activity evidence.
func SuggestGapFill(
	ctx context.Context,
	client *Client,
	model string,
	categories []string,
	gap GapContext,
	evidence []EvidencePayload,
	local bool,
) (category string, description string, err error) {
	if len(categories) == 0 {
		return "", "", fmt.Errorf("no categories configured")
	}

	payload := evidence
	if !local {
		payload = minimizeEvidence(evidence)
	}

	gapJSON, err := json.Marshal(gap)
	if err != nil {
		return "", "", err
	}
	evidenceJSON, err := json.Marshal(payload)
	if err != nil {
		return "", "", err
	}

	categoryList := strings.Join(categories, ", ")
	systemPrompt := "You suggest timesheet entries for uncovered work intervals. Reply with JSON only: {\"category\":\"<one category from the list>\",\"description\":\"<short note>\"}."
	userPrompt := fmt.Sprintf(
		"Categories: %s\nGap: %s\nActivity evidence: %s\nJSON:",
		categoryList,
		string(gapJSON),
		string(evidenceJSON),
	)

	reply, err := client.ChatCompletion(ctx, model, systemPrompt, userPrompt)
	if err != nil {
		return "", "", err
	}

	var parsed struct {
		Category    string `json:"category"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(extractJSONObject(reply)), &parsed); err != nil {
		return "", "", fmt.Errorf("parse gap-fill reply: %w", err)
	}

	category = strings.Trim(parsed.Category, `"' `)
	description = strings.TrimSpace(parsed.Description)
	if category == "" {
		return "", "", fmt.Errorf("model returned empty category")
	}

	for _, name := range categories {
		if strings.EqualFold(category, name) {
			return name, description, nil
		}
	}

	lowerCategory := strings.ToLower(category)
	for _, name := range categories {
		if strings.Contains(lowerCategory, strings.ToLower(name)) {
			return name, description, nil
		}
	}

	return "", "", fmt.Errorf("model returned unknown category %q", category)
}

func minimizeEvidence(items []EvidencePayload) []EvidencePayload {
	out := make([]EvidencePayload, 0, len(items))
	for _, item := range items {
		out = append(out, EvidencePayload{
			Provider: item.Provider,
			Kind:     item.Kind,
			Summary:  item.Summary,
		})
	}
	return out
}

func extractJSONObject(reply string) string {
	reply = strings.TrimSpace(reply)
	if strings.HasPrefix(reply, "```") {
		reply = strings.TrimPrefix(reply, "```json")
		reply = strings.TrimPrefix(reply, "```")
		reply = strings.TrimSuffix(reply, "```")
		reply = strings.TrimSpace(reply)
	}
	start := strings.Index(reply, "{")
	end := strings.LastIndex(reply, "}")
	if start >= 0 && end > start {
		return reply[start : end+1]
	}
	return reply
}
