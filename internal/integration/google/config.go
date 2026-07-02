package google

import (
	"github.com/dylanbr0wn/clockr/internal/integration/oauth"
	"github.com/dylanbr0wn/clockr/internal/service"
)

const (
	apiBaseURL        = "https://www.googleapis.com/calendar/v3"
	scopeCalendarRead = "https://www.googleapis.com/auth/calendar.readonly"
	authURL           = "https://accounts.google.com/o/oauth2/auth"
	tokenURL          = "https://oauth2.googleapis.com/token"
	calendarListPath  = "/users/me/calendarList"
	eventsListPath    = "/calendars/%s/events"
)

// OAuthConfig builds reusable Google OAuth settings for the calendar provider.
func OAuthConfig(clientID, clientSecret string) oauth.ProviderConfig {
	return oauth.ProviderConfig{
		Provider:     service.ProviderGoogle,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AuthURL:      authURL,
		TokenURL:     tokenURL,
		Scopes:       []string{scopeCalendarRead},
	}
}
