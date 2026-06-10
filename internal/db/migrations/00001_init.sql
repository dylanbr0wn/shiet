-- +goose Up
-- +goose StatementBegin

-- Mirrors DESIGN.md "Persistence & period model" + multi-source calendars + TZ segments.
-- All instants stored as RFC3339 UTC text. Date-only fields stored as YYYY-MM-DD text.
-- Booleans stored as INTEGER 0/1.

-- User-defined, free-form categories. Exactly one may be the default gap category.
CREATE TABLE category (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    name           TEXT    NOT NULL UNIQUE,
    is_default_gap INTEGER NOT NULL DEFAULT 0 CHECK (is_default_gap IN (0, 1)),
    created_at     TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
-- At most one default gap category.
CREATE UNIQUE INDEX idx_category_one_default_gap
    ON category (is_default_gap) WHERE is_default_gap = 1;

-- A live, editable working record for a pay period. Boundaries deterministic from
-- cadence + anchor; row created lazily on first open of a date range.
CREATE TABLE period (
    id                   INTEGER PRIMARY KEY AUTOINCREMENT,
    start_date           TEXT    NOT NULL,                 -- YYYY-MM-DD, inclusive
    end_date             TEXT    NOT NULL,                 -- YYYY-MM-DD, inclusive
    cadence              TEXT    NOT NULL CHECK (cadence IN ('weekly', 'bi-weekly', 'semi-monthly', 'monthly')),
    anchor_date          TEXT    NOT NULL,                 -- YYYY-MM-DD
    target_hours_per_day REAL    NOT NULL DEFAULT 8,
    last_synced_at       TEXT,                             -- RFC3339 UTC, null until first sync
    created_at           TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (start_date, end_date)
);

-- Date-anchored TZ segments covering a period (handles split-location / DST).
-- Default = one segment at the device TZ when the period is created.
CREATE TABLE tz_segment (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id           INTEGER NOT NULL REFERENCES period (id) ON DELETE CASCADE,
    effective_from_date TEXT    NOT NULL,                  -- YYYY-MM-DD
    iana_tz             TEXT    NOT NULL,                  -- e.g. "America/Toronto"
    UNIQUE (period_id, effective_from_date)
);

-- Account-level calendar scope. Defaults to primary only; user toggles `selected`.
-- Deselecting soft-hides events (event.active) rather than deleting them.
CREATE TABLE calendar (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    google_calendar_id  TEXT    NOT NULL UNIQUE,
    name                TEXT    NOT NULL,
    is_primary          INTEGER NOT NULL DEFAULT 0 CHECK (is_primary IN (0, 1)),
    selected            INTEGER NOT NULL DEFAULT 0 CHECK (selected IN (0, 1)),
    default_category_id INTEGER REFERENCES category (id) ON DELETE SET NULL,
    created_at          TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- Synced facts: imported calendar events. Never carry user decisions (those live in
-- overlay). Cross-calendar identity = ical_uid; per-calendar id = google_event_id.
-- Recurring occurrence = recurring_event_id (series) + instance_id (originalStartTime).
CREATE TABLE event (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id          INTEGER NOT NULL REFERENCES period (id) ON DELETE CASCADE,
    calendar_id        INTEGER NOT NULL REFERENCES calendar (id) ON DELETE CASCADE,
    google_event_id    TEXT    NOT NULL,
    instance_id        TEXT    NOT NULL DEFAULT '',        -- originalStartTime for recurring occurrence
    recurring_event_id TEXT    NOT NULL DEFAULT '',        -- series id
    ical_uid           TEXT    NOT NULL DEFAULT '',        -- cross-calendar identity
    title              TEXT    NOT NULL,
    description        TEXT    NOT NULL DEFAULT '',
    location           TEXT    NOT NULL DEFAULT '',
    organizer          TEXT    NOT NULL DEFAULT '',        -- organizer email
    attendees          TEXT    NOT NULL DEFAULT '[]',      -- JSON array of attendee objects
    status             TEXT    NOT NULL DEFAULT '',        -- accepted | declined | tentative | needsAction
    all_day            INTEGER NOT NULL DEFAULT 0 CHECK (all_day IN (0, 1)),
    start_utc          TEXT,                               -- RFC3339 UTC instant (null for all-day)
    end_utc            TEXT,
    start_date         TEXT,                               -- literal date for all-day (YYYY-MM-DD)
    end_date           TEXT,
    original_tz        TEXT    NOT NULL DEFAULT '',        -- IANA TZ from Google
    active             INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)), -- soft-hide on deselect
    source_hash        TEXT    NOT NULL,                   -- hash of synced fields for change detection
    created_at         TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at         TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (period_id, calendar_id, google_event_id, instance_id)
);
CREATE INDEX idx_event_period   ON event (period_id);
CREATE INDEX idx_event_ical_uid ON event (ical_uid) WHERE ical_uid <> '';
CREATE INDEX idx_event_gid      ON event (google_event_id);

-- User decisions, re-attached to facts by google_event_id (+ instance_id). Survives
-- re-sync. kind distinguishes the decision type for a given event occurrence.
CREATE TABLE overlay (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id        INTEGER NOT NULL REFERENCES period (id) ON DELETE CASCADE,
    google_event_id  TEXT    NOT NULL,
    instance_id      TEXT    NOT NULL DEFAULT '',
    category_id      INTEGER REFERENCES category (id) ON DELETE SET NULL,
    resolved_overlap TEXT,                                 -- JSON: interval -> owning category
    note             TEXT    NOT NULL DEFAULT '',
    kind             TEXT    NOT NULL DEFAULT 'category' CHECK (kind IN ('category', 'overlap', 'status')),
    created_at       TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at       TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (period_id, google_event_id, instance_id, kind)
);
CREATE INDEX idx_overlay_period ON overlay (period_id);

-- Entries filling uncovered intervals (gaps), plus manual blocks / overtime extensions.
CREATE TABLE gap_fill (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id   INTEGER NOT NULL REFERENCES period (id) ON DELETE CASCADE,
    day         TEXT    NOT NULL,                          -- YYYY-MM-DD (active-segment day)
    start_utc   TEXT    NOT NULL,                          -- RFC3339 UTC
    end_utc     TEXT    NOT NULL,
    category_id INTEGER REFERENCES category (id) ON DELETE SET NULL,
    note        TEXT    NOT NULL DEFAULT '',
    source      TEXT    NOT NULL DEFAULT 'gap' CHECK (source IN ('gap', 'manual', 'extend')),
    created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);
CREATE INDEX idx_gap_fill_period_day ON gap_fill (period_id, day);

-- Categorization memory. Match key = recurringEventId (series) | normalized title (+organizer).
CREATE TABLE memory (
    match_key   TEXT    PRIMARY KEY,
    category_id INTEGER NOT NULL REFERENCES category (id) ON DELETE CASCADE,
    hits        INTEGER NOT NULL DEFAULT 1,
    updated_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- Immutable finalized snapshots. Each finalize writes a new version; prior versions kept.
CREATE TABLE submission (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id    INTEGER NOT NULL REFERENCES period (id) ON DELETE CASCADE,
    version      INTEGER NOT NULL,
    finalized_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    frozen_blob  TEXT    NOT NULL,                         -- JSON snapshot of events + overlays
    UNIQUE (period_id, version)
);

-- Conflicts surfaced by re-sync / dedup that need explicit user resolution.
CREATE TABLE review_item (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id   INTEGER NOT NULL REFERENCES period (id) ON DELETE CASCADE,
    kind        TEXT    NOT NULL CHECK (kind IN (
                    'new_in_gap', 'title_changed', 'deleted_categorized',
                    'dedup_ambiguous', 'overlap', 'tentative', 'all_day')),
    event_id    INTEGER REFERENCES event (id) ON DELETE CASCADE,
    payload     TEXT    NOT NULL DEFAULT '{}',             -- JSON context for the decision
    status      TEXT    NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved', 'dismissed')),
    created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    resolved_at TEXT
);
CREATE INDEX idx_review_item_period_status ON review_item (period_id, status);

-- Non-secret app config (BYOM endpoint, privacy field toggles, event-rule defaults).
-- Secrets (API keys, OAuth tokens) live in the OS keychain, never here.
CREATE TABLE app_setting (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,                              -- JSON-encoded value
    updated_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS app_setting;
DROP TABLE IF EXISTS review_item;
DROP TABLE IF EXISTS submission;
DROP TABLE IF EXISTS memory;
DROP TABLE IF EXISTS gap_fill;
DROP TABLE IF EXISTS overlay;
DROP TABLE IF EXISTS event;
DROP TABLE IF EXISTS calendar;
DROP TABLE IF EXISTS tz_segment;
DROP TABLE IF EXISTS period;
DROP TABLE IF EXISTS category;
-- +goose StatementEnd
