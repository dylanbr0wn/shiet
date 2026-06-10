-- name: GetSetting :one
SELECT value FROM app_setting WHERE key = ?;

-- name: ListSettings :many
SELECT * FROM app_setting ORDER BY key;

-- name: SetSetting :exec
INSERT INTO app_setting (key, value)
VALUES (?, ?)
ON CONFLICT (key) DO UPDATE SET
    value      = excluded.value,
    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now');
