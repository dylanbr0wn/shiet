#!/usr/bin/env bash
# Release validation for public Google Calendar OAuth packaging (DYL-86).
# Always runs no-shared-secret checks. Optionally:
#   RUN_BROKER_SMOKE=1  — HTTPS health check against the broker
#   RUN_OAUTH_TESTS=0   — skip focused Go auth tests (default: run)
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

fail() {
  echo "release-oauth-check: $*" >&2
  exit 1
}

echo "==> Checking public builds do not embed a shared Google client secret"

google_cfg="internal/integration/google/config.go"
if [[ ! -f "$google_cfg" ]]; then
  fail "missing $google_cfg"
fi

if ! grep -q 'defaultDesktopClientID[[:space:]]*=[[:space:]]*""' "$google_cfg"; then
  fail "defaultDesktopClientID must be empty in $google_cfg"
fi
if ! grep -q 'defaultDesktopClientSecret[[:space:]]*=[[:space:]]*""' "$google_cfg"; then
  fail "defaultDesktopClientSecret must be empty in $google_cfg"
fi

# Fail on non-empty google.client_secret in tracked example/dev config.
# Empty quotes / bare keys are allowed.
check_yaml_secret() {
  local file="$1"
  [[ -f "$file" ]] || return 0
  local line
  while IFS= read -r line; do
    case "$line" in
      *client_secret:*)
        local value
        value="$(printf '%s\n' "$line" | sed -E 's/^[[:space:]]*client_secret:[[:space:]]*//; s/[[:space:]]+#.*$//; s/^["'\''](.*)["'\'']$/\1/; s/^[[:space:]]+//; s/[[:space:]]+$//')"
        if [[ -n "$value" && "$value" != '""' && "$value" != "''" ]]; then
          fail "$file embeds non-empty client_secret (public builds must not ship a shared secret)"
        fi
        ;;
    esac
  done <"$file"
}

check_yaml_secret config.example.yaml
check_yaml_secret shiet.yaml

if [[ -d build/bin ]]; then
  # Soft binary scan: fail only on obvious embedded credential assignment markers.
  if command -v rg >/dev/null 2>&1; then
    if rg -a -n 'defaultDesktopClientSecret[[:space:]]*=[[:space:]]*"[^"]+"' build/bin 2>/dev/null \
      | rg -v 'defaultDesktopClientSecret[[:space:]]*=[[:space:]]*""' >/dev/null 2>&1; then
      fail "build/bin appears to embed a non-empty defaultDesktopClientSecret"
    fi
  fi
fi

echo "no-shared-secret check passed"

if [[ "${RUN_OAUTH_TESTS:-1}" != "0" ]]; then
  echo "==> Broker connect/refresh smoke (Go): desktop BrokerFlow + broker HTTP API"
  go test ./internal/config/ ./internal/integration/google/ ./internal/broker/... -count=1
fi

if [[ "${RUN_BROKER_SMOKE:-0}" == "1" ]]; then
  broker_base="${BROKER_BASE_URL:-https://auth.shiet.app}"
  broker_base="${broker_base%/}"
  echo "==> Deployed broker liveness: ${broker_base}/healthz"
  if ! curl -fsS --max-time 15 "${broker_base}/healthz" >/dev/null; then
    fail "broker healthz failed at ${broker_base}/healthz"
  fi
  if ! curl -fsS --max-time 15 "${broker_base}/readyz" >/dev/null; then
    fail "broker readyz failed at ${broker_base}/readyz"
  fi
  echo "deployed broker liveness passed"
else
  echo "deployed broker HTTP liveness skipped (set RUN_BROKER_SMOKE=1 to enable)"
fi

echo "Manual release checklist (when credentials/network allow):"
echo "  - Connect Google via Settings in a public build (broker mode)"
echo "  - Confirm token refresh after expiry / forced 401 path"
echo "release-oauth-check: ok"
