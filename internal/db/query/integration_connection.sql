-- name: ListIntegrationConnections :many
SELECT id, provider, account_label, account_id, scopes, status, connected_at, updated_at
FROM integration_connection
ORDER BY provider, account_label;

-- name: ListIntegrationConnectionsByProvider :many
SELECT id, provider, account_label, account_id, scopes, status, connected_at, updated_at
FROM integration_connection
WHERE provider = ?
ORDER BY account_label;

-- name: GetIntegrationConnection :one
SELECT id, provider, account_label, account_id, scopes, status, connected_at, updated_at
FROM integration_connection
WHERE provider = ? AND account_id = ?;

-- name: UpsertIntegrationConnection :one
INSERT INTO integration_connection (
    provider, account_label, account_id, scopes, status, connected_at, updated_at
) VALUES (
    ?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'), strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
)
ON CONFLICT (provider, account_id) DO UPDATE SET
    account_label = excluded.account_label,
    scopes        = excluded.scopes,
    status        = excluded.status,
    updated_at    = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
RETURNING id, provider, account_label, account_id, scopes, status, connected_at, updated_at;

-- name: UpdateIntegrationConnectionStatus :exec
UPDATE integration_connection
SET status = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE provider = ? AND account_id = ?;

-- name: DeleteIntegrationConnection :exec
DELETE FROM integration_connection
WHERE provider = ? AND account_id = ?;
