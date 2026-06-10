-- name: ListCalendars :many
SELECT * FROM calendar ORDER BY is_primary DESC, name;

-- name: ListSelectedCalendars :many
SELECT * FROM calendar WHERE selected = 1 ORDER BY is_primary DESC, name;

-- name: GetCalendarByGoogleID :one
SELECT * FROM calendar WHERE google_calendar_id = ?;

-- name: UpsertCalendar :one
INSERT INTO calendar (google_calendar_id, name, is_primary)
VALUES (?, ?, ?)
ON CONFLICT (google_calendar_id) DO UPDATE SET
    name = excluded.name,
    is_primary = excluded.is_primary
RETURNING *;

-- name: SetCalendarSelected :exec
UPDATE calendar SET selected = ? WHERE id = ?;

-- name: SetCalendarDefaultCategory :exec
UPDATE calendar SET default_category_id = ? WHERE id = ?;
