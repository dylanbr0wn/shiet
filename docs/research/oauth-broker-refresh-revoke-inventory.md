# OAuth broker Refresh / Revoke RPC inventory

_Research date: 2026-07-22_  
_Linear: [DYL-185](https://linear.app/dylans-apps/issue/DYL-185/inventory-refreshrevoke-broker-rpcs)_

## Question and scope

What are the exact `OAuthBrokerService.RefreshToken` / `RevokeToken` Connect RPC shapes (request/response fields, supported providers, error codes) for Google, GitHub, Slack, and Bitbucket — and what do local wrappers strip, add, or remap — so oauth strategy hooks can be sized without guessing RPC shapes?

**In scope:** broker Connect RPCs `RefreshToken` / `RevokeToken` (`brokerv1`), broker server enforcement, desktop `broker_flow.go` wrappers, `oauth.Capabilities`, and error mapping on those paths.

**Out of scope:** app-level “RefreshGitHubRepos” / channel resource sync RPCs; Authorize/StartAuthorization/ExchangeHandoff beyond shared pieces that refresh/revoke wrappers reuse.

Primary sources: `proto/shiet/broker/v1/oauth_broker.proto`, `internal/broker/httpapi/operations.go`, provider `broker_flow.go` files, `internal/integration/oauth/{providers,broker_flow,provider}.go`, `internal/broker/codes/codes.go`, `internal/broker/httpapi/errors.go`.

---

## Summary table

| Provider | Registry `Capabilities` | Broker `RefreshToken` | Broker `RevokeToken` | Local wrapper Refresh | Local wrapper Revoke | Notable gaps |
|----------|-------------------------|------------------------|----------------------|-----------------------|----------------------|--------------|
| Google | Refresh ✅ Revoke ✅ | Supported | Supported (`refresh_token` only) | Yes — full adapter | Yes — refresh credential | Scopes forwarded to Google form; Application metadata sent |
| GitHub | Refresh ❌ Revoke ✅ | `operation_not_supported` → Connect `Unimplemented` | Supported (`access_token` only) | **None** | Yes — access credential | No refresh path by design (OAuth App tokens) |
| Slack | Refresh ❌ Revoke ✅ | `operation_not_supported` → Connect `Unimplemented` | Supported (`access_token` only; Slack API) | **None** | Yes — access credential | Same refresh gap as GitHub |
| Bitbucket | Refresh ✅ Revoke ❌ | Supported (token endpoint via provider creds) | **No dedicated revoke**; non-Google branch currently falls through to Slack revoke helper if access_token supplied | Yes — near-Google copy | **None**; Disconnect deletes keychain only | Scopes on refresh request ignored by broker; platform default `"desktop"` |

Sources: capabilities in [`internal/integration/oauth/providers.go`](../../internal/integration/oauth/providers.go) L31, L45, L60, L73; refresh gate in [`internal/broker/httpapi/operations.go`](../../internal/broker/httpapi/operations.go) L226–228; revoke branching L336–364; wrappers under `internal/integration/{google,github,slack,bitbucket}/broker_flow.go`.

---

## Broker RPC contract (proto)

Service and RPCs ([`proto/shiet/broker/v1/oauth_broker.proto`](../../proto/shiet/broker/v1/oauth_broker.proto) L9–14):

```text
service OAuthBrokerService {
  rpc StartAuthorization(...)
  rpc ExchangeHandoff(...)
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc RevokeToken(RevokeTokenRequest) returns (RevokeTokenResponse);
}
```

### Provider enum

| Enum | Value |
|------|-------|
| `PROVIDER_UNSPECIFIED` | 0 |
| `PROVIDER_GOOGLE` | 1 |
| `PROVIDER_GITHUB` | 2 |
| `PROVIDER_SLACK` | 3 |
| `PROVIDER_BITBUCKET` | 4 |

Source: same proto L16–22.

### Shared messages

**`ApplicationMetadata`** (L24–27):

| Field | Type | Number |
|-------|------|--------|
| `app_version` | string | 1 |
| `platform` | string | 2 |

**`TokenMaterial`** (L29–34):

| Field | Type | Number |
|-------|------|--------|
| `access_token` | string | 1 |
| `refresh_token` | string | 2 |
| `token_type` | string | 3 |
| `expiry` | `google.protobuf.Timestamp` | 4 |

**`BrokerErrorDetail`** (L90–92): `code` string field 1 — attached as Connect error detail by the broker ([`internal/broker/httpapi/connect.go`](../../internal/broker/httpapi/connect.go) L48–54).

### `RefreshToken`

**Request** ([`oauth_broker.proto`](../../proto/shiet/broker/v1/oauth_broker.proto) L66–71):

| Field | Type | Number | Notes |
|-------|------|--------|-------|
| `provider` | `Provider` | 1 | Required for routing |
| `refresh_token` | string | 2 | Empty → `refresh_token_required` |
| `scopes` | `repeated string` | 3 | Broker applies only for Google (joined with spaces into form `scope`) |
| `application` | `ApplicationMetadata` | 4 | Optional; `app_version` used for kill-switch |

**Response** (L73–75):

| Field | Type | Number |
|-------|------|--------|
| `token` | `TokenMaterial` | 1 |

### `RevokeToken`

**Request** (L77–84):

| Field | Type | Number | Notes |
|-------|------|--------|-------|
| `provider` | `Provider` | 1 | |
| `credential` oneof | | | Exactly one credential kind expected per provider |
| → `refresh_token` | string | 2 | Google path |
| → `access_token` | string | 3 | GitHub / Slack path |
| `reason` | string | 4 | Logged; not validated for enum |

**Response** (L86–88):

| Field | Type | Number |
|-------|------|--------|
| `revoked` | bool | 1 |

---

## Broker server enforcement (facts)

### RefreshToken (`BrokerService.refreshToken`)

Source: [`internal/broker/httpapi/operations.go`](../../internal/broker/httpapi/operations.go) L225–319.

1. **Hard reject:** `PROVIDER_GITHUB` or `PROVIDER_SLACK` → `operation_not_supported` (L226–228). Connect maps that to `CodeUnimplemented` ([`errors.go`](../../internal/broker/httpapi/errors.go) L31–32). Covered for GitHub by [`TestConnectRefreshTokenSupportsGoogleOnly`](../../internal/broker/httpapi/connect_test.go) L254–260.
2. Unrecognized / unconfigured provider → `provider_not_configured`.
3. Kill switches: `RefreshDisabled` → `refresh_disabled`; `AppVersionDisabled(app_version)` → `app_version_disabled`.
4. Empty `refresh_token` → `refresh_token_required`.
5. Form always sets `grant_type=refresh_token` + `refresh_token`.
6. **Scopes:** only `PROVIDER_GOOGLE` copies `req.Scopes` into form `scope` (L261–267). Bitbucket (and any other non-Google supported provider) **does not** set scope on the outbound form even if the client sent `scopes` (L268–275).
7. Token POST: Google via `postGoogleToken`; otherwise `postProviderToken` (Bitbucket).
8. Error mapping:
   - Provider token error with Google `invalid_grant` → `invalid_refresh_token` (L292–297).
   - Any Bitbucket refresh failure → `invalid_refresh_token` (L299–302).
   - Other Google failures → `google_token_refresh_failed` (L304–306).
9. Success: `TokenMaterial` with `Expiry = now + expires_in seconds` (L314–318). Empty provider `token_type` defaulted to `"Bearer"` on the broker (L308–311).

### RevokeToken (`BrokerService.revokeToken`)

Source: [`operations.go`](../../internal/broker/httpapi/operations.go) L322–369.

1. Provider must resolve + be configured; else `provider_not_configured`.
2. Rate-limit key: Google uses IP bucket alone; others use `provider|ip` (L327–330).
3. **Google** (L336–350):
   - Requires non-empty `refresh_token` and empty access side of the oneof (`GetAccessToken() != ""` rejects) → else `refresh_token_required`.
   - Calls `revokeGoogleToken`. Already-revoked/`invalid_token` treated as success (`revoked: true`) (L342–346).
   - Other failures → `google_revoke_failed`.
4. **Non-Google** (L351–364):
   - Requires non-empty `access_token` and empty refresh side → else `access_token_required`.
   - If `provider == github` → `revokeGitHubToken`; **else** → `revokeSlackToken` (L356–363).
   - Consequence: a Bitbucket `RevokeToken` with `access_token` would hit the Slack revoke helper and Slack revoke URL (`https://slack.com/api/auth.revoke`), not a Bitbucket API. There is no `bitbucket_revoke_failed` code. Desktop Bitbucket does not call revoke today.

Upstream helpers: Google/GitHub/Slack revoke implementations in [`server.go`](../../internal/broker/httpapi/server.go) L701–777+.

---

## Capability flags / unsupported sentinels

[`Capabilities`](../../internal/integration/oauth/provider.go) L25–30:

```go
type Capabilities struct {
	Refresh bool
	Revoke  bool
}
```

Registered values ([`providers.go`](../../internal/integration/oauth/providers.go)):

| Provider ID | Refresh | Revoke | RevokeURL in descriptor |
|-------------|---------|--------|-------------------------|
| `google` | true | true | `https://oauth2.googleapis.com/revoke` |
| `github` | false | true | `https://api.github.com` |
| `slack` | false | true | `https://slack.com/api/auth.revoke` |
| `bitbucket` | true | false | *(empty — not set)* |

Tests assert Google refresh+revoke and GitHub no-refresh ([`provider_test.go`](../../internal/integration/oauth/provider_test.go) L24–36).

Broker unsupported refresh sentinel string: `codes.OperationNotSupported` = `"operation_not_supported"` ([`codes.go`](../../internal/broker/codes/codes.go) L37).

---

## Shared oauth package (Authorize + error helpers)

### Authorize path (shared; refresh/revoke stay provider-local)

[`internal/integration/oauth/broker_flow.go`](../../internal/integration/oauth/broker_flow.go):

- Comment L43–45: provider packages keep provider-specific refresh/revoke; Authorize is shared.
- Sends `ApplicationMetadata` on StartAuthorization / ExchangeHandoff (L214, L233).
- Defaults: `appVersion` → `"dev"` (L310–314); `platform` → `runtime.GOOS + "-" + runtime.GOARCH` (L317–321).
- Shared sentinels (L34–41): `ErrBrokerUnavailable`, `ErrBrokerRejected`, `ErrHandoffReplay`, `ErrHandoffExpired`, `ErrHandoffStateMismatch`, `ErrHandoffVerifier`.
- Provider packages re-export these for Authorize/UI `errors.Is`.

### `BrokerErrorCode` / `mapBrokerRPCError`

| Symbol | Location | Role |
|--------|----------|------|
| `BrokerErrorCode(err)` | [`broker_flow.go`](../../internal/integration/oauth/broker_flow.go) L273–288 | Extracts `BrokerErrorDetail.code` from Connect error details; falls back to Connect message |
| `mapBrokerRPCError(err, op)` | L253–271 | **Authorize-only** (start/handoff). Maps handoff codes + `rate_limited` / `auth_disabled` / `app_version_disabled`; Unavailable/Internal → `ErrBrokerUnavailable` |

There is **no** shared unbranded `MapBrokerRPCError` for refresh/revoke. Google and Bitbucket each define private `(*BrokerFlow).mapBrokerRPCError`; GitHub/Slack inline a thinner mapping on Revoke only.

### Legacy JSON companion types

[`internal/integration/oauth/protocol.go`](../../internal/integration/oauth/protocol.go) still documents older REST-shaped structs (`BrokerRefreshRequest` “Google-only today”, `BrokerRevokeRequest` with parallel refresh/access fields). Live application transport for these ops is Connect only ([`docs/oauth-broker.md`](../../docs/oauth-broker.md) L250–254). Strategy sizing should use the **proto** fields above, not the legacy JSON field names (`scope` singular vs proto `scopes`).

---

## Per-provider sections

### Google

**Files:** [`internal/integration/google/broker_flow.go`](../../internal/integration/google/broker_flow.go), Disconnect/refresh wiring in [`provider.go`](../../internal/integration/google/provider.go).

#### RefreshToken wrapper

- Method: `RefreshToken(ctx, refreshToken string, scopes []string) (secrets.Token, error)` (L73–111).
- **Adds:**
  - Hardcodes `Provider: PROVIDER_GOOGLE`.
  - Copies `scopes` into request (`append` clone).
  - Sets `Application` via `appVersion()` / `platform()` (defaults `"dev"`, `GOOS-GOARCH`).
- **Validates locally:** empty base URL → `config.ErrBrokerConfig`; empty refresh → `ErrInvalidRefreshToken`.
- **Response remapping to `secrets.Token`:**
  - Requires non-nil token + non-empty `access_token`; else `ErrBrokerUnavailable`.
  - Default `token_type` to `"Bearer"` if empty.
  - If response `refresh_token` empty, **reuses input** refresh token (rotation-safe fallback) (L101–104).
  - Sets `Expiry` from proto timestamp `AsTime()`.
  - Does **not** set `CredentialSource` on the returned token.
- Caller: `brokerTokenRefresher` passes `p.Config.Scopes` ([`provider.go`](../../internal/integration/google/provider.go) L298–313).

#### RevokeToken wrapper

- Method: `Revoke(ctx, refreshToken string) error` (L115–138).
- **Adds:** `Credential: RefreshToken{...}`, `Reason: "user_disconnect"`.
- Does **not** send `Application` metadata.
- Requires `response.Msg.Revoked == true`; else `ErrBrokerRejected`.
- Disconnect: best-effort revoke when broker mode + refresh present; always deletes keychain ([`provider.go`](../../internal/integration/google/provider.go) L111–141). Skip revoke when no refresh token (tested in `TestDisconnect_brokerModeSkipsRevokeWithoutRefreshToken`).

#### Error mapping (Google-local)

[`mapBrokerRPCError`](../../internal/integration/google/broker_flow.go) L165–196 maps (among others used on refresh/revoke):

| Detail code | Desktop sentinel / wrap |
|-------------|-------------------------|
| `invalid_refresh_token` | `ErrInvalidRefreshToken` (“reconnect Google Calendar”) |
| `rate_limited` | `ErrBrokerRejected` |
| `auth_disabled` | `ErrBrokerRejected` |
| `refresh_disabled` | `ErrBrokerRejected` |
| `app_version_disabled` | `ErrBrokerRejected` |
| Handoff codes (`handoff_*`) | Mapped (Authorize-oriented; unlikely on refresh/revoke) |
| Connect Unavailable/Internal | `ErrBrokerUnavailable` |
| Other non-empty code | `ErrBrokerRejected` with code string |

Google-specific local sentinel: `ErrInvalidRefreshToken` (L31).

---

### GitHub

**Files:** [`internal/integration/github/broker_flow.go`](../../internal/integration/github/broker_flow.go), [`provider.go`](../../internal/integration/github/provider.go).

#### Refresh

- **No** desktop `RefreshToken` method.
- Registry: `Capabilities.Refresh = false`.
- Broker: immediate `operation_not_supported` / Connect `Unimplemented`.
- Comment documents OAuth App tokens as non-expiring / no refresh (broker_flow.go L27–29; also [`docs/oauth-broker.md`](../../docs/oauth-broker.md) L276–278).

#### RevokeToken wrapper

- Method: `Revoke(ctx, accessToken string) error` (L57–87).
- **Adds:** `PROVIDER_GITHUB`, `Credential: AccessToken{...}`, `Reason: "user_disconnect"`.
- No `Application` metadata.
- Requires `revoked == true`.
- **Error mapping (minimal):** extracts code via `oauth.BrokerErrorCode`; Unavailable/Internal → `ErrBrokerUnavailable`; else `ErrBrokerRejected` with code string. **Does not** special-case `github_revoke_failed`, `rate_limited`, etc.
- Disconnect only revokes when `CredentialSource == broker` and access token present ([`provider.go`](../../internal/integration/github/provider.go) L180–187). PATs are not broker-revoked.

`TokenRevoker` interface parameter is named `accessToken` ([`provider.go`](../../internal/integration/github/provider.go) L53–56).

---

### Slack

**Files:** [`internal/integration/slack/broker_flow.go`](../../internal/integration/slack/broker_flow.go), [`provider.go`](../../internal/integration/slack/provider.go).

#### Refresh

- **No** desktop refresh wrapper; `Capabilities.Refresh = false`; broker same GitHub hard-reject.

#### RevokeToken wrapper

- Structurally identical to GitHub revoke (L55–85): `PROVIDER_SLACK`, access_token credential, `Reason: "user_disconnect"`, same minimal error mapping, requires `revoked == true`.
- Disconnect gate: broker credential source + access token ([`provider.go`](../../internal/integration/slack/provider.go) L174–181).

Broker upstream uses Slack `auth.revoke` with Bearer access token ([`server.go`](../../internal/broker/httpapi/server.go) L701–737). Failure detail code: `slack_revoke_failed`.

---

### Bitbucket

**Files:** [`internal/integration/bitbucket/broker_flow.go`](../../internal/integration/bitbucket/broker_flow.go), [`provider.go`](../../internal/integration/bitbucket/provider.go).

#### RefreshToken wrapper

- Method signature matches Google: `RefreshToken(ctx, refreshToken, scopes) (secrets.Token, error)` (L58–96).
- Same response remapping (Bearer default, preserve refresh if rotation omitted, no `CredentialSource`).
- **Adds** `PROVIDER_BITBUCKET`, scopes clone, `Application` metadata.
- **Difference vs Google `platform()`:** Bitbucket defaults platform to literal `"desktop"` when unset (L113–117), not `GOOS-GOARCH`.
- Caller still passes `p.Config.Scopes` into the RPC ([`provider.go`](../../internal/integration/bitbucket/provider.go) L401–417), but broker **ignores** those scopes for non-Google refresh (see server enforcement above).
- Local sentinel: `ErrInvalidRefreshToken` (broker_flow.go L27).

#### Revoke

- **No** `Revoke` method on Bitbucket `BrokerFlow`.
- `Capabilities.Revoke = false`; no `RevokeURL` in registry.
- `Disconnect` deletes DB rows + keychain only — **no broker revoke call** ([`provider.go`](../../internal/integration/bitbucket/provider.go) L146–171).

#### Error mapping (Bitbucket-local, refresh)

[`mapBrokerRPCError`](../../internal/integration/bitbucket/broker_flow.go) L120–137:

| Detail code | Mapping |
|-------------|---------|
| `invalid_refresh_token` | `ErrInvalidRefreshToken` |
| `rate_limited` | `ErrBrokerRejected` |
| `auth_disabled`, `refresh_disabled`, `app_version_disabled` | Single branded “Bitbucket auth temporarily unavailable” |
| Unavailable/Internal | `ErrBrokerUnavailable` |
| Other code | `ErrBrokerRejected` + code |

Unlike Google, does not map handoff detail codes on this path.

---

## Error-code surface on refresh/revoke paths

### Codes defined (broker package)

From [`internal/broker/codes/codes.go`](../../internal/broker/codes/codes.go) L7–38, those **emitted or specially relevant** on refresh/revoke:

| Code constant | String | Typical Connect code ([`errors.go`](../../internal/broker/httpapi/errors.go)) | Produced on |
|---------------|--------|-----------------------------------------------------------------------------|-------------|
| `RefreshTokenRequired` | `refresh_token_required` | InvalidArgument | Empty refresh (refresh RPC or Google revoke credential mismatch) |
| `AccessTokenRequired` | `access_token_required` | InvalidArgument | Non-Google revoke credential mismatch |
| `InvalidRefreshToken` | `invalid_refresh_token` | InvalidArgument | Google `invalid_grant`; all Bitbucket refresh failures |
| `GoogleTokenRefreshFailed` | `google_token_refresh_failed` | Unavailable | Non-`invalid_grant` Google refresh failure |
| `GoogleRevokeFailed` | `google_revoke_failed` | Unavailable | Google revoke failure (not already-revoked) |
| `GitHubRevokeFailed` | `github_revoke_failed` | Unavailable | GitHub revoke failure |
| `SlackRevokeFailed` | `slack_revoke_failed` | Unavailable | Slack revoke failure |
| `OperationNotSupported` | `operation_not_supported` | Unimplemented | GitHub/Slack refresh |
| `ProviderNotConfigured` | `provider_not_configured` | Unavailable | Missing provider config |
| `RateLimited` | `rate_limited` | ResourceExhausted | Refresh/revoke rate limits |
| `RefreshDisabled` | `refresh_disabled` | FailedPrecondition | Refresh kill switch |
| `AppVersionDisabled` | `app_version_disabled` | FailedPrecondition | Version kill switch on refresh |
| `AuthDisabled` | `auth_disabled` | FailedPrecondition | (mapped by Google/Bitbucket clients; authorize-oriented kill switch) |

There is **no** `bitbucket_revoke_failed` / `bitbucket_token_refresh_failed` code.

### Who maps what today

| Layer | Refresh | Revoke |
|-------|---------|--------|
| Shared `oauth.mapBrokerRPCError` | Not used | Not used |
| Google `mapBrokerRPCError` | Yes (rich) | Yes (rich) |
| Bitbucket `mapBrokerRPCError` | Yes (medium) | N/A |
| GitHub / Slack inline | N/A | Minimal (Unavailable/Internal vs generic rejected) |

Desktop clients do **not** special-case `google_token_refresh_failed`, `*_revoke_failed`, or `operation_not_supported` into dedicated sentinels (those fall through to generic `ErrBrokerRejected` / Connect code checks where present).

---

## Shape differences that matter for strategy-hook sizing (facts only)

1. **Credential kind for revoke is provider-split:** Google = refresh_token oneof arm; GitHub/Slack = access_token arm; Bitbucket has no desktop revoke and registry `Revoke=false`. Proto is a single RPC with a `oneof`.
2. **Refresh support is not universal:** GitHub/Slack broker returns `operation_not_supported`; only Google + Bitbucket refresh server-side. Registry flags already encode this.
3. **`scopes` on RefreshTokenRequest are Google-meaningful only** at the broker; Bitbucket wrappers still send scopes but the broker omits them from the token form.
4. **`Application` metadata:** sent on Google/Bitbucket refresh; **not** sent on any current revoke wrapper; Authorize always sends it via shared `oauth.BrokerFlow`.
5. **`reason`:** all current revoke wrappers hardcode `"user_disconnect"`; broker only logs it.
6. **Response enrichment on refresh:** wrappers default `token_type`, preserve prior refresh token when rotation omitted, convert expiry; strip proto nesting into `secrets.Token`; do not round-trip `CredentialSource`.
7. **Revoke success contract:** wrappers require `revoked=true`; Google broker can return success for already-revoked tokens.
8. **Platform default inconsistency:** Google/shared Authorize use `GOOS-GOARCH`; Bitbucket refresh helper defaults `"desktop"`.
9. **Error mapper duplication:** three styles (shared authorize mapper, Google rich refresh/revoke mapper, Bitbucket refresh-only mapper, GitHub/Slack minimal revoke). No single refresh/revoke strategy mapper exists yet.
10. **Local method shapes already diverge:** `RefreshToken(ctx, refreshToken, scopes)` vs absent; `Revoke(ctx, refreshToken)` vs `Revoke(ctx, accessToken)` vs absent — same Go signature name `Revoke` but different credential semantics by package.
11. **Stale docs vs code:** [`docs/oauth-broker.md`](../../docs/oauth-broker.md) L268–271 still describes Refresh as Google-only and Revoke as Google refresh / GitHub access; code also refreshes Bitbucket and revokes Slack. Prefer `operations.go` + proto for contracts.

---

## Key source files

| Path | Why cited |
|------|-----------|
| `proto/shiet/broker/v1/oauth_broker.proto` | Canonical Connect message/RPC field layout |
| `internal/broker/httpapi/operations.go` | Provider support gates + form/credential rules |
| `internal/broker/httpapi/errors.go` | Detail code → Connect code |
| `internal/broker/httpapi/connect.go` | Detail attachment |
| `internal/broker/httpapi/connect_test.go` | Observed Unimplemented refresh; credential validation |
| `internal/broker/httpapi/server.go` | Provider revoke HTTP implementations |
| `internal/broker/codes/codes.go` | Stable detail code strings |
| `internal/integration/oauth/providers.go` | Capability flags |
| `internal/integration/oauth/provider.go` | `Capabilities` type |
| `internal/integration/oauth/broker_flow.go` | Shared Authorize + `BrokerErrorCode` / authorize `mapBrokerRPCError` |
| `internal/integration/oauth/protocol.go` | Legacy JSON shapes (not live Connect) |
| `internal/integration/google/broker_flow.go` | Google refresh+revoke adapter + mapper |
| `internal/integration/github/broker_flow.go` | GitHub revoke-only adapter |
| `internal/integration/slack/broker_flow.go` | Slack revoke-only adapter |
| `internal/integration/bitbucket/broker_flow.go` | Bitbucket refresh-only adapter |
| `internal/integration/{google,github,slack,bitbucket}/provider.go` | Disconnect/refresh call sites and credential kinds |
