package ai

// Endpoint is a known or user-supplied OpenAI-compatible API base URL.
type Endpoint struct {
	Name    string   `json:"name"`
	BaseURL string   `json:"baseUrl"`
	Local   bool     `json:"local"`
	Running bool     `json:"running"`
	Models  []string `json:"models,omitempty"`
}

// ValidationResult summarizes a connectivity + privacy check.
type ValidationResult struct {
	OK      bool   `json:"ok"`
	Local   bool   `json:"local"`
	Verdict string `json:"verdict"`
	Message string `json:"message"`
}

// EventContext is the event payload sent to a categorization model.
type EventContext struct {
	Title           string   `json:"title"`
	Description     string   `json:"description,omitempty"`
	Location        string   `json:"location,omitempty"`
	Organizer       string   `json:"organizer,omitempty"`
	AttendeeDomains []string `json:"attendeeDomains,omitempty"`
	Duration        string   `json:"duration,omitempty"`
}

// PrivacyFields controls which event fields may be sent to cloud models.
type PrivacyFields struct {
	Title       bool `json:"title"`
	Attendees   bool `json:"attendees"`
	Description bool `json:"description"`
	Location    bool `json:"location"`
}
