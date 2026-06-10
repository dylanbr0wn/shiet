#!/usr/bin/env bash
# Regenerate type-safe query code from internal/db/query/*.sql + the migration
# schema. Run after editing any .sql query or migration.
set -euo pipefail
cd "$(dirname "$0")/.."
exec go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate
