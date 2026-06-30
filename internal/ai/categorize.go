package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// SuggestCategory asks the model to pick one category name for an event.
func SuggestCategory(
	ctx context.Context,
	client *Client,
	model string,
	categories []string,
	event EventContext,
	local bool,
	privacy PrivacyFields,
) (string, error) {
	if len(categories) == 0 {
		return "", fmt.Errorf("no categories configured")
	}

	payload := event
	if !local {
		payload = minimizeEvent(event, privacy)
	}

	eventJSON, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	categoryList := strings.Join(categories, ", ")
	systemPrompt := "You categorize calendar events for a timesheet app. Reply with exactly one category name from the provided list and nothing else."
	userPrompt := fmt.Sprintf(
		"Categories: %s\nEvent: %s\nCategory:",
		categoryList,
		string(eventJSON),
	)

	reply, err := client.ChatCompletion(ctx, model, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	reply = strings.Trim(reply, `"' `)
	for _, category := range categories {
		if strings.EqualFold(reply, category) {
			return category, nil
		}
	}

	// Accept a reply that merely contains a known category name.
	lowerReply := strings.ToLower(reply)
	for _, category := range categories {
		if strings.Contains(lowerReply, strings.ToLower(category)) {
			return category, nil
		}
	}

	return "", fmt.Errorf("model returned unknown category %q", reply)
}

func minimizeEvent(event EventContext, privacy PrivacyFields) EventContext {
	out := EventContext{}
	if privacy.Title {
		out.Title = event.Title
	}
	if privacy.Description {
		out.Description = event.Description
	}
	if privacy.Location {
		out.Location = event.Location
	}
	if privacy.Attendees {
		out.AttendeeDomains = event.AttendeeDomains
	}
	out.Duration = event.Duration
	return out
}
