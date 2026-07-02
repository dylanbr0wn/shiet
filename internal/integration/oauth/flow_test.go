package oauth_test

import (
	"testing"

	"github.com/dylanbr0wn/clockr/internal/integration/oauth"
)

func TestParseCallback(t *testing.T) {
	code, err := oauth.ParseCallback(
		"http://127.0.0.1:1234/oauth/callback?code=abc&state=xyz",
		"xyz",
	)
	if err != nil || code != "abc" {
		t.Fatalf("ParseCallback: code=%q err=%v", code, err)
	}

	_, err = oauth.ParseCallback(
		"http://127.0.0.1:1234/oauth/callback?code=abc&state=bad",
		"xyz",
	)
	if err == nil {
		t.Fatal("expected state mismatch")
	}
}

func TestProviderOAuth2Config(t *testing.T) {
	cfg := oauth.ProviderConfig{
		Provider: "google",
		ClientID: "client-id",
		AuthURL:  "https://accounts.google.com/o/oauth2/auth",
		TokenURL: "https://oauth2.googleapis.com/token",
		Scopes:   []string{"calendar.readonly"},
	}.OAuth2Config("http://127.0.0.1:8080/oauth/callback")

	if cfg.ClientID != "client-id" {
		t.Fatalf("client id: %q", cfg.ClientID)
	}
	if cfg.RedirectURL != "http://127.0.0.1:8080/oauth/callback" {
		t.Fatalf("redirect: %q", cfg.RedirectURL)
	}
	if len(cfg.Scopes) != 1 || cfg.Scopes[0] != "calendar.readonly" {
		t.Fatalf("scopes: %#v", cfg.Scopes)
	}
}
