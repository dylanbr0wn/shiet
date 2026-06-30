-- name: ListOverlaysForPeriod :many
SELECT * FROM overlay WHERE period_id = ?;

-- name: GetOverlay :one
SELECT * FROM overlay
WHERE period_id = ? AND provider = ? AND external_id = ? AND instance_id = ? AND kind = ?;

-- name: UpsertOverlay :one
INSERT INTO overlay (
    period_id, provider, external_id, instance_id, category_id, resolved_overlap, note, kind
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (period_id, provider, external_id, instance_id, kind) DO UPDATE SET
    category_id      = excluded.category_id,
    resolved_overlap = excluded.resolved_overlap,
    note             = excluded.note,
    updated_at       = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
RETURNING *;

-- name: DeleteOverlay :exec
DELETE FROM overlay WHERE id = ?;
