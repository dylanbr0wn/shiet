package log

import (
	"encoding/json"
	"strings"
)

// Redacted is the replacement value for scrubbed secrets and tokens.
const Redacted = "[redacted]"

// sensitiveAttrNames are never logged with their values (ADR-0001).
var sensitiveAttrNames = map[string]struct{}{
	"access_token":            {},
	"refresh_token":           {},
	"client_secret":           {},
	"handoff_code":            {},
	"authorization":           {},
	"authorization_code":      {},
	"code":                    {},
	"token":                   {},
	"pkce_verifier":           {},
	"handoff_verifier":        {},
	"encrypted_token_payload": {},
}

// SensitiveKey reports whether a log field name must always be redacted.
func SensitiveKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	if _, ok := sensitiveAttrNames[key]; ok {
		return true
	}
	return strings.HasSuffix(key, "_token") || strings.HasSuffix(key, "_secret")
}

// LooksLikeSecret reports whether a string value looks like Google OAuth
// token material (access tokens ya29.*, refresh tokens 1//*).
func LooksLikeSecret(v string) bool {
	v = strings.TrimSpace(v)
	if len(v) < 20 {
		return false
	}
	lower := strings.ToLower(v)
	return strings.HasPrefix(lower, "ya29.") || strings.HasPrefix(lower, "1//")
}

// ShouldRedact reports whether a key/value pair must be replaced with Redacted.
func ShouldRedact(key, value string) bool {
	if SensitiveKey(key) {
		return true
	}
	return LooksLikeSecret(value)
}

// RedactJSONLine rewrites a single JSON object log line, replacing sensitive
// string field values with Redacted. Non-JSON input is returned unchanged.
func RedactJSONLine(line []byte) []byte {
	trim := bytesTrimRightSpace(line)
	newline := len(line) > len(trim)
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(trim, &obj); err != nil {
		return line
	}
	changed := false
	for k, raw := range obj {
		var s string
		if err := json.Unmarshal(raw, &s); err != nil {
			continue
		}
		if !ShouldRedact(k, s) {
			continue
		}
		obj[k] = json.RawMessage(`"` + Redacted + `"`)
		changed = true
	}
	if !changed {
		return line
	}
	out, err := json.Marshal(obj)
	if err != nil {
		return line
	}
	if newline {
		out = append(out, '\n')
	}
	return out
}

func bytesTrimRightSpace(b []byte) []byte {
	i := len(b)
	for i > 0 {
		c := b[i-1]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			break
		}
		i--
	}
	return b[:i]
}
