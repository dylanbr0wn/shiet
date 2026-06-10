-- name: ListCategories :many
SELECT * FROM category ORDER BY name;

-- name: GetCategory :one
SELECT * FROM category WHERE id = ?;

-- name: GetDefaultGapCategory :one
SELECT * FROM category WHERE is_default_gap = 1;

-- name: CreateCategory :one
INSERT INTO category (name, is_default_gap)
VALUES (?, ?)
RETURNING *;

-- name: RenameCategory :exec
UPDATE category SET name = ? WHERE id = ?;

-- name: ClearDefaultGap :exec
UPDATE category SET is_default_gap = 0 WHERE is_default_gap = 1;

-- name: SetDefaultGap :exec
UPDATE category SET is_default_gap = 1 WHERE id = ?;

-- name: DeleteCategory :exec
DELETE FROM category WHERE id = ?;
