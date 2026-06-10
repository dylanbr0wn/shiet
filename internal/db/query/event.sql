-- name: ListEventsForPeriod :many
SELECT * FROM event WHERE period_id = ? AND active = 1 ORDER BY start_utc, start_date;

-- name: ListAllEventsForPeriod :many
SELECT * FROM event WHERE period_id = ? ORDER BY start_utc, start_date;

-- name: GetEvent :one
SELECT * FROM event WHERE id = ?;

-- name: ListEventsByIcalUID :many
SELECT * FROM event WHERE period_id = ? AND ical_uid = ? AND ical_uid <> '';

-- name: UpsertEvent :one
-- Re-sync entry point: insert a fact, or update mutable synced fields on re-pull.
-- Never touches user decisions (those live in overlay), and preserves `active`.
INSERT INTO event (
    period_id, calendar_id, google_event_id, instance_id, recurring_event_id,
    ical_uid, title, description, location, organizer, attendees, status,
    all_day, start_utc, end_utc, start_date, end_date, original_tz, source_hash
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (period_id, calendar_id, google_event_id, instance_id) DO UPDATE SET
    recurring_event_id = excluded.recurring_event_id,
    ical_uid           = excluded.ical_uid,
    title              = excluded.title,
    description        = excluded.description,
    location           = excluded.location,
    organizer          = excluded.organizer,
    attendees          = excluded.attendees,
    status             = excluded.status,
    all_day            = excluded.all_day,
    start_utc          = excluded.start_utc,
    end_utc            = excluded.end_utc,
    start_date         = excluded.start_date,
    end_date           = excluded.end_date,
    original_tz        = excluded.original_tz,
    source_hash        = excluded.source_hash,
    updated_at         = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
RETURNING *;

-- name: SetEventActiveByCalendar :exec
-- Soft-hide / restore all events from a calendar when it is deselected / reselected.
UPDATE event SET active = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE calendar_id = ?;

-- name: DeleteEvent :exec
DELETE FROM event WHERE id = ?;
