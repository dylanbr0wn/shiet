-- name: GetMemory :one
SELECT * FROM memory WHERE match_key = ?;

-- name: ListMemory :many
SELECT * FROM memory ORDER BY hits DESC;

-- name: RememberCategory :one
-- Train memory from a user correction; bump hit count on repeat.
INSERT INTO memory (match_key, category_id, hits)
VALUES (?, ?, 1)
ON CONFLICT (match_key) DO UPDATE SET
    category_id = excluded.category_id,
    hits        = memory.hits + 1,
    updated_at  = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
RETURNING *;

-- name: ForgetMemory :exec
DELETE FROM memory WHERE match_key = ?;
