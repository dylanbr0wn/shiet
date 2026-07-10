#!/bin/bash
# Simple production build for current platform

set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "Running OAuth release pre-check (source defaults)..."
"${root}/scripts/release-oauth-check.sh"

echo "Building for production..."
wails build -clean
echo "Build complete! Check build/bin/"

echo "Re-running OAuth release check against build artifacts..."
RUN_OAUTH_TESTS=0 "${root}/scripts/release-oauth-check.sh"
