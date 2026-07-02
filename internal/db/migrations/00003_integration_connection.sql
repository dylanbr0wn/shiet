-- +goose Up
-- +goose StatementBegin

-- Non-secret metadata for connected integration accounts. OAuth tokens live in the
-- OS keychain (see internal/integration/secrets), never in SQLite.
CREATE TABLE integration_connection (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    provider      TEXT    NOT NULL,
    account_label TEXT    NOT NULL,
    account_id    TEXT    NOT NULL,
    scopes        TEXT    NOT NULL DEFAULT '[]', -- JSON array of scope strings
    status        TEXT    NOT NULL DEFAULT 'connected'
        CHECK (status IN ('connected', 'needs_reauth', 'disconnected')),
    connected_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (provider, account_id)
);

CREATE INDEX idx_integration_connection_provider
    ON integration_connection (provider);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS integration_connection;

-- +goose StatementEnd
