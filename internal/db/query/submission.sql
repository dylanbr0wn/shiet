-- name: ListSubmissions :many
SELECT * FROM submission WHERE period_id = ? ORDER BY version DESC;

-- name: GetLatestSubmission :one
SELECT * FROM submission WHERE period_id = ? ORDER BY version DESC LIMIT 1;

-- name: NextSubmissionVersion :one
SELECT COALESCE(MAX(version), 0) + 1 AS next_version FROM submission WHERE period_id = ?;

-- name: CreateSubmission :one
INSERT INTO submission (period_id, version, frozen_blob)
VALUES (?, ?, ?)
RETURNING *;
