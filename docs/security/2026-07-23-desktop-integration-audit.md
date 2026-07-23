# Security Audit: Desktop Integration Layer

**Date**: 2026-07-23
**Scope**: OAuth flows, secret management, and token handling in `internal/integration/`
**Auditor**: Automated cloud agent

## Executive Summary

Reviewed 16 files across the OAuth, secrets, connection, httpclient, and provider
packages. Found **11 new findings** ranging from high to informational severity.
Two previously reported findings (Bitbucket pagination SSRF, port-based AI endpoint
classification bypass) are excluded.

---

## Findings

### FINDING-01 — Broker Base URL Accepts Plaintext HTTP

| Field | Value |
|-------|-------|
| **Severity** | High |
| **Files** | `internal/integration/google/broker_flow.go:74`, `internal/integration/github/broker_flow.go:62`, `internal/integration/slack/broker_flow.go:60`, `internal/integration/bitbucket/broker_flow.go:59`, `internal/integration/oauth/broker_flow.go:76` |
| **CWE** | CWE-319 (Cleartext Transmission of Sensitive Information) |

**Description**: Every broker flow constructs a Connect RPC client from the
configured `BaseURL` without verifying the scheme is `https`. If
`SHIET_GOOGLE_BROKER_BASE_URL` (or any provider variant) is set to an `http://`
URL, all broker RPCs — `StartAuthorization`, `ExchangeHandoff`, `RefreshToken`,
`RevokeToken` — transmit refresh tokens, access tokens, PKCE challenges, and
handoff codes in plaintext.

**Attack path**: An attacker on the same network segment (e.g., coffee-shop Wi-Fi)
performs passive sniffing or ARP spoofing. The refresh token from `RefreshToken`
calls gives persistent access to the user's Google Calendar / GitHub / Slack /
Bitbucket data.

**Impact**: Full account takeover for any connected integration if broker URL is
misconfigured to HTTP. Refresh tokens grant long-lived access.

**Recommendation**: Validate that `BaseURL` uses the `https` scheme before
constructing the broker client. Reject `http` except when the host is
`127.0.0.1` or `localhost` (for local development).

---

### FINDING-02 — Config-Driven Token URL Override Redirects Auth Code + PKCE Verifier

| Field | Value |
|-------|-------|
| **Severity** | Medium |
| **Files** | `internal/integration/oauth/flow.go:93-98,182-183`, `internal/integration/oauth/authorize.go:116-120` |
| **CWE** | CWE-601 (URL Redirection to Untrusted Site) |

**Description**: `ProviderConfig.TokenURL` and `ProviderConfig.AuthURL` can be
overridden via the layered config system (`shiet.yaml`, env vars). When
`TokenURL` is overridden, `ExchangeAuthorizationCode` sends the authorization
code **and** the PKCE code verifier to the overridden endpoint. An attacker who
can write a config file to any search path (`./shiet.yaml`,
`~/.config/shiet/config.yaml`) can redirect the token exchange to a server they
control.

**Attack path**:
1. Attacker places a malicious `shiet.yaml` in the working directory (e.g., via
   a cloned repo or shared filesystem).
2. User launches shiet, which loads the poisoned `token_url`.
3. User initiates OAuth. The authorization code + PKCE verifier are POSTed to
   the attacker's server.
4. Attacker replays the code + verifier at the real provider token endpoint,
   obtaining access and refresh tokens.

**Impact**: Full OAuth token theft. PKCE is defeated because the verifier is
sent to the attacker alongside the code.

**Recommendation**: Pin token URLs to the provider registry and reject
config-level overrides in production builds. If overrides are needed for
development, gate them behind a `--dev` flag or `SHIET_DEV=1` env var.

---

### FINDING-03 — Unbounded Retry-After Delay Enables Provider-Side DoS

| Field | Value |
|-------|-------|
| **Severity** | Medium |
| **Files** | `internal/integration/httpclient/client.go:150-165` |
| **CWE** | CWE-400 (Uncontrolled Resource Consumption) |

**Description**: `retryAfter()` converts the `Retry-After` header to a
`time.Duration` with no upper bound. A malicious or compromised provider API
can return `Retry-After: 999999` (≈11.5 days), causing the HTTP client goroutine
to sleep indefinitely. While context cancellation eventually unblocks it, the
user-facing operation hangs until the parent context times out (up to 30 s for
the default HTTP timeout, but context may be longer for sync operations).

**Attack path**: A compromised or attacker-controlled API endpoint (e.g., via
BaseURL override) returns 429 with an extreme `Retry-After`. The desktop app
freezes the sync operation.

**Impact**: Denial of service for integration sync; user must force-quit or wait
for context timeout.

**Recommendation**: Cap `retryAfter` at a reasonable maximum (e.g., 120 seconds).

---

### FINDING-04 — Log Redaction Heuristic Misses Non-Google Token Formats

| Field | Value |
|-------|-------|
| **Severity** | Medium |
| **Files** | `internal/log/redact.go:37-44` |
| **CWE** | CWE-532 (Insertion of Sensitive Information into Log File) |

**Description**: `LooksLikeSecret()` only recognizes Google token prefixes
(`ya29.*` for access tokens, `1//*` for refresh tokens). GitHub tokens
(`gho_*`, `ghp_*`, `ghs_*`), Slack tokens (`xoxp-*`, `xoxb-*`, `xoxa-*`), and
Bitbucket tokens have different formats and would not be caught by this value
heuristic.

Token values are normally protected by `SensitiveKey()` matching the field name
(e.g., `access_token`). However, if a token value appears under an unexpected
key name (e.g., embedded in a larger JSON structure, or a third-party library
logging under a non-standard key), `LooksLikeSecret()` is the last line of
defense — and it only works for Google.

**Impact**: GitHub/Slack/Bitbucket tokens could appear in log files under
non-standard field names.

**Recommendation**: Add known prefixes for all supported providers:
`gho_`, `ghp_`, `ghs_`, `github_pat_`, `xoxp-`, `xoxb-`, `xoxa-`, `xoxe-`.

---

### FINDING-05 — Broker Success Page Meta-Refresh Bypasses html/template URL Sanitization

| Field | Value |
|-------|-------|
| **Severity** | Medium |
| **Files** | `internal/oauthpages/assets/success.html:8,32` |
| **CWE** | CWE-79 (Improper Neutralization of Input During Web Page Generation) |

**Description**: The broker success page uses `{{.HandoffURL}}` in two contexts:

```html
<meta http-equiv="refresh" content="0;url={{.HandoffURL}}">
<a class="button" href="{{.HandoffURL}}">Open shiet</a>
```

Go's `html/template` applies context-aware escaping. For the `href` attribute,
it recognizes URLs and sanitizes dangerous schemes (e.g., `javascript:`).
However, inside the `content` attribute of a `<meta>` tag, the template engine
treats it as a plain string attribute — it does **not** apply URL-context
sanitization. A `HandoffURL` containing `javascript:alert(1)` or
`data:text/html,...` would be HTML-entity-escaped but not scheme-blocked in the
meta-refresh context.

The `href` attribute is safe (template blocks `javascript:` there). The risk
is specifically in the meta-refresh auto-redirect.

In practice, the `HandoffURL` is built server-side from validated inputs
(`validateDesktopHandoffRedirect` ensures `http://127.0.0.1` scheme+host), so
exploitation requires a compromised broker. The fallback
`fallbackSuccessPage()` uses `html.EscapeString()` which also doesn't block
dangerous URL schemes.

**Impact**: If a broker is compromised or returns a crafted handoff URL, the
meta-refresh could redirect to an attacker-controlled page or execute script.

**Recommendation**: Explicitly validate `HandoffURL` scheme (allow only `http`
and the app's custom scheme) before passing it to the template. Alternatively,
use `template.URL` type only after validation.

---

### FINDING-06 — TOCTOU Race in Broker State Validation

| Field | Value |
|-------|-------|
| **Severity** | Low |
| **Files** | `internal/integration/oauth/broker_flow.go:98-153` |
| **CWE** | CWE-367 (Time-of-check Time-of-use Race Condition) |

**Description**: The broker flow's loopback server starts accepting connections
before `expectedState` is set. The handler at line 117 skips state validation
when `wantState` is empty:

```go
if wantState != "" && state != wantState {
```

Between server start (line 138) and state assignment (line 152), a local process
that discovers the random loopback port could send a callback with any
`broker_state` value and it would pass validation.

The injected `handoff_code` would fail at the broker's `ExchangeHandoff` (the
broker validates it server-side), so this cannot lead to token theft. However,
the `codeCh` channel (buffer 1) would be consumed, and the legitimate callback
arriving later would block, causing the flow to time out.

**Impact**: Local DoS — an attacker monitoring loopback ports during the narrow
window could force the OAuth flow to fail.

**Recommendation**: Initialize `expectedState` to a sentinel value and reject
callbacks while it is unset, rather than allowing any state when empty.

---

### FINDING-07 — OAuth State Compared With Non-Constant-Time Equality

| Field | Value |
|-------|-------|
| **Severity** | Low |
| **Files** | `internal/integration/oauth/flow.go:120`, `internal/integration/oauth/broker_flow.go:118` |
| **CWE** | CWE-208 (Observable Timing Discrepancy) |

**Description**: OAuth state parameters are compared with Go's `!=` operator:

```go
if r.URL.Query().Get("state") != state {
```

This is a non-constant-time string comparison. In theory, a timing side-channel
could leak the state value byte-by-byte.

**Practical exploitability**: Very low. The comparison happens on a localhost
HTTP server, and the state is a 32-byte random value (256 bits of entropy).
Network jitter on localhost is orders of magnitude larger than the timing
difference from string comparison short-circuiting.

**Impact**: Theoretical timing side-channel; not practically exploitable.

**Recommendation**: Use `subtle.ConstantTimeCompare` as a defense-in-depth
measure. The effort is minimal and eliminates the theoretical concern.

---

### FINDING-08 — Expired Tokens Sent to Provider API Before Refresh

| Field | Value |
|-------|-------|
| **Severity** | Low |
| **Files** | `internal/integration/httpclient/client.go:51-61` |
| **CWE** | CWE-324 (Use of a Key Past its Expiration Date) |

**Description**: `Client.Do()` retrieves the stored token and immediately sends
it in the `Authorization` header without checking `token.Expiry`. If the token
is expired, the request fails with 401, triggering a refresh-then-retry cycle.

This means every API call with an expired token:
1. Sends the expired (but still sensitive) token over the network unnecessarily.
2. Incurs an extra round-trip to the provider API.
3. Counts against rate limits.

**Impact**: Minor — expired tokens typically cannot be used by an interceptor,
but sending them unnecessarily increases exposure surface.

**Recommendation**: Check `token.Expiry` before sending the request. If expired
(with a small buffer, e.g., 30 seconds), refresh proactively.

---

### FINDING-09 — No Pagination Limit on Calendar/Channel Sync Loops

| Field | Value |
|-------|-------|
| **Severity** | Low |
| **Files** | `internal/integration/google/provider.go:162-199`, `internal/integration/slack/provider.go:221-267` |
| **CWE** | CWE-400 (Uncontrolled Resource Consumption) |

**Description**: `SyncCalendars` and `SyncChannels` loop until `pageToken` /
`cursor` is empty, with no maximum page count. If a provider API (or a
test-injected BaseURL) returns infinite pagination tokens, the loop runs
indefinitely.

**Impact**: Resource exhaustion (CPU, memory, API quota) via infinite pagination.
Requires a compromised or attacker-controlled API endpoint.

**Recommendation**: Add a maximum page count (e.g., 100 pages) consistent with
the Bitbucket evidence collector which already caps at `maxCommitPages`.

---

### FINDING-10 — Raw API Response Bodies in Error Messages

| Field | Value |
|-------|-------|
| **Severity** | Low |
| **Files** | `internal/integration/github/provider.go:324-326`, `internal/integration/slack/provider.go:319`, `internal/integration/bitbucket/provider.go:356`, `internal/integration/google/provider.go:280` |
| **CWE** | CWE-209 (Generation of Error Message Containing Sensitive Information) |

**Description**: When provider API calls fail, the raw HTTP response body is
included in the error message:

```go
return fmt.Errorf("invalid personal access token (github api %s: %s)",
    userPath, strings.TrimSpace(string(body)))
```

Provider error responses sometimes include request metadata, internal error IDs,
or partial request echoes. These errors propagate up through the call chain and
may reach logging or UI layers. While the redacting logger scrubs known-sensitive
keys, these error strings bypass key-based redaction.

**Impact**: Potential information disclosure of provider-internal details in logs
or error UI.

**Recommendation**: Extract only the structured error code/message from provider
responses; do not embed the full response body in error strings. Use
`log.Reason()` for structured log entries.

---

### FINDING-11 — Exchange Error Details Rendered in Loopback Callback Page

| Field | Value |
|-------|-------|
| **Severity** | Low |
| **Files** | `internal/integration/oauth/flow.go:186-190` |
| **CWE** | CWE-209 (Generation of Error Message Containing Sensitive Information) |

**Description**: When the token exchange fails, the full error message
(including `describeExchangeError` output) is rendered in the browser callback
page:

```go
sendCallbackResult(resultCh, callbackResult{
    status:  http.StatusBadGateway,
    message: exchangeErr.Error(),
})
```

The `describeExchangeError` function produces messages containing internal
config key names (`google.client_secret`, `SHIET_GOOGLE_CLIENT_SECRET`,
`github.auth_mode`), deployment hints, and wrapped provider error details.

The page is served on localhost and uses `html/template` (auto-escaped, no XSS),
but the content reveals internal configuration structure to any local observer.

**Impact**: Information disclosure of internal configuration layout; no direct
exploitation path.

**Recommendation**: Display a generic user-friendly error in the callback page.
Log the detailed error server-side for debugging.

---

## Positive Observations

The following security controls are correctly implemented:

- **PKCE S256**: Properly implemented with `crypto/rand` + SHA-256 + base64url.
- **Loopback binding**: OAuth callback servers correctly bind to `127.0.0.1`
  (not `0.0.0.0`).
- **State parameter entropy**: 32 bytes from `crypto/rand` (256 bits).
- **Authorization URL validation**: Broker-returned auth URLs are validated
  against allowlisted hosts and paths before opening in the browser.
- **Desktop handoff redirect validation**: Broker validates `http://127.0.0.1`
  scheme, no userinfo, no query/fragment.
- **Log redaction framework**: Comprehensive sensitive-key list with redacting
  writer; `Reason()` function prevents raw error strings in structured logs.
- **No InsecureSkipVerify**: No TLS verification bypass found anywhere.
- **Token store abstraction**: Clean separation between keyring (production) and
  memory (test) stores.
- **Handoff replay prevention**: Broker marks handoffs as used, preventing
  replay attacks.
- **PKCE verifier isolation**: Verifier never appears in URLs or browser
  history; only sent in the POST body to the token endpoint.
