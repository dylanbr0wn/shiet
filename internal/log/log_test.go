package log

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestShouldRedactScrubsTokensAndSecrets(t *testing.T) {
	cases := []struct {
		key, value string
		want       bool
	}{
		{"event", "refresh_failed", false},
		{"reason", "invalid_grant", false},
		{"refresh_token", "1//secret-refresh-token-value", true},
		{"client_secret", "super-secret-value", true},
		{"handoff_code", "abc123handoff", true},
		{"access_token", "ya29.a0AfH6SMB-example-token", true},
		{"authorization_code", "4/0AeanS...", true},
		{"pkce_verifier", "verifier-material", true},
		{"handoff_verifier", "handoff-verifier", true},
		{"custom_token", "anything", true},
		{"api_secret", "anything", true},
		{"note", "ya29.a0AfH6SMB-example-token", true}, // value shape
		{"note", "1//should-redact-this-refresh", true},
		{"note", "short", false},
	}
	for _, tc := range cases {
		got := ShouldRedact(tc.key, tc.value)
		if got != tc.want {
			t.Fatalf("ShouldRedact(%q, %q)=%v want %v", tc.key, tc.value, got, tc.want)
		}
	}
}

func TestRedactJSONLine(t *testing.T) {
	in := []byte(`{"event":"refresh_failed","refresh_token":"1//secret-refresh-token-value","client_secret":"super-secret-value","handoff_code":"abc123handoff","access_token":"ya29.a0AfH6SMB-example-token","reason":"invalid_grant"}` + "\n")
	out := RedactJSONLine(in)
	var obj map[string]string
	if err := json.Unmarshal(bytes.TrimSpace(out), &obj); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, out)
	}
	if obj["event"] != "refresh_failed" {
		t.Fatalf("event: %q", obj["event"])
	}
	if obj["reason"] != "invalid_grant" {
		t.Fatalf("reason: %q", obj["reason"])
	}
	for _, key := range []string{"refresh_token", "client_secret", "handoff_code", "access_token"} {
		if obj[key] != Redacted {
			t.Fatalf("%s: got %q want %s", key, obj[key], Redacted)
		}
	}
	if !bytes.HasSuffix(out, []byte("\n")) {
		t.Fatalf("expected trailing newline preserved")
	}
}

func TestLoggerRedactsInJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf)
	logger.Info().
		Str("surface", "refresh").
		Str("refresh_token", "1//should-not-appear-in-logs").
		Str("client_secret", "google-client-secret").
		Str("handoff_code", "handoff-secret-code").
		Str("outcome", "failure").
		Msg("broker_event")

	got := buf.String()
	for _, leak := range []string{
		"1//should-not-appear",
		"google-client-secret",
		"handoff-secret-code",
	} {
		if strings.Contains(got, leak) {
			t.Fatalf("secret leaked (%q): %s", leak, got)
		}
	}
	if !strings.Contains(got, Redacted) {
		t.Fatalf("expected redaction marker: %s", got)
	}
	if !strings.Contains(got, `"outcome":"failure"`) {
		t.Fatalf("expected safe field preserved: %s", got)
	}
	if !strings.Contains(got, `"surface":"refresh"`) {
		t.Fatalf("expected safe field preserved: %s", got)
	}
}

func TestSensitiveKeyCaseInsensitive(t *testing.T) {
	if !SensitiveKey("Refresh_Token") {
		t.Fatal("expected Refresh_Token sensitive")
	}
	if !SensitiveKey("CLIENT_SECRET") {
		t.Fatal("expected CLIENT_SECRET sensitive")
	}
}
