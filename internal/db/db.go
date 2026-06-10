// Package db owns the SQLite connection, schema migrations, and (via the sqlc
// subpackage) type-safe queries for Clockr's local store.
package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // pure-Go driver, registers under the name "sqlite"
)

// DefaultPath returns the on-disk location of the user's Clockr database,
// under the OS-specific user config dir (e.g. ~/Library/Application Support
// on macOS). The CLOCKR_DB env var overrides it (used by dev tooling).
func DefaultPath() (string, error) {
	if p := os.Getenv("CLOCKR_DB"); p != "" {
		return p, nil
	}
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("locate user config dir: %w", err)
	}
	return filepath.Join(cfg, "clockr", "clockr.db"), nil
}

// Open opens (creating parent dirs as needed) the SQLite database at path and
// applies the connection pragmas Clockr relies on: WAL journaling, enforced
// foreign keys, and a busy timeout. It does NOT run migrations — call Migrate.
func Open(path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create db dir: %w", err)
		}
	}

	dsn := fmt.Sprintf(
		"file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=synchronous(NORMAL)",
		path,
	)
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return conn, nil
}

// OpenDefault opens the database at DefaultPath.
func OpenDefault() (*sql.DB, error) {
	path, err := DefaultPath()
	if err != nil {
		return nil, err
	}
	return Open(path)
}
