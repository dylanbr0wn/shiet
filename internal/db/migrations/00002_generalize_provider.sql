-- +goose Up
-- +goose StatementBegin

-- Generalize Google-specific identifiers to provider + external_id.
-- Existing rows are migrated as provider='google'.

-- calendar: google_calendar_id -> (provider, external_id)
CREATE TABLE calendar_new (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    provider            TEXT    NOT NULL,
    external_id         TEXT    NOT NULL,
    name                TEXT    NOT NULL,
    is_primary          INTEGER NOT NULL DEFAULT 0 CHECK (is_primary IN (0, 1)),
    selected            INTEGER NOT NULL DEFAULT 0 CHECK (selected IN (0, 1)),
    default_category_id INTEGER REFERENCES category (id) ON DELETE SET NULL,
    created_at          TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (provider, external_id)
);

INSERT INTO calendar_new (
    id, provider, external_id, name, is_primary, selected, default_category_id, created_at
)
SELECT id, 'google', google_calendar_id, name, is_primary, selected, default_category_id, created_at
FROM calendar;

DROP TABLE calendar;
ALTER TABLE calendar_new RENAME TO calendar;

-- event: google_event_id -> (provider, external_id)
CREATE TABLE event_new (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id          INTEGER NOT NULL REFERENCES period (id) ON DELETE CASCADE,
    calendar_id        INTEGER NOT NULL REFERENCES calendar (id) ON DELETE CASCADE,
    provider           TEXT    NOT NULL,
    external_id        TEXT    NOT NULL,
    instance_id        TEXT    NOT NULL DEFAULT '',
    recurring_event_id TEXT    NOT NULL DEFAULT '',
    ical_uid           TEXT    NOT NULL DEFAULT '',
    title              TEXT    NOT NULL,
    description        TEXT    NOT NULL DEFAULT '',
    location           TEXT    NOT NULL DEFAULT '',
    organizer          TEXT    NOT NULL DEFAULT '',
    attendees          TEXT    NOT NULL DEFAULT '[]',
    status             TEXT    NOT NULL DEFAULT '',
    all_day            INTEGER NOT NULL DEFAULT 0 CHECK (all_day IN (0, 1)),
    start_utc          TEXT,
    end_utc            TEXT,
    start_date         TEXT,
    end_date           TEXT,
    original_tz        TEXT    NOT NULL DEFAULT '',
    active             INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
    source_hash        TEXT    NOT NULL,
    created_at         TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at         TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (period_id, calendar_id, external_id, instance_id)
);

INSERT INTO event_new (
    id, period_id, calendar_id, provider, external_id, instance_id, recurring_event_id,
    ical_uid, title, description, location, organizer, attendees, status, all_day,
    start_utc, end_utc, start_date, end_date, original_tz, active, source_hash,
    created_at, updated_at
)
SELECT
    id, period_id, calendar_id, 'google', google_event_id, instance_id, recurring_event_id,
    ical_uid, title, description, location, organizer, attendees, status, all_day,
    start_utc, end_utc, start_date, end_date, original_tz, active, source_hash,
    created_at, updated_at
FROM event;

DROP TABLE event;
ALTER TABLE event_new RENAME TO event;

CREATE INDEX idx_event_period ON event (period_id);
CREATE INDEX idx_event_ical_uid ON event (ical_uid) WHERE ical_uid <> '';
CREATE INDEX idx_event_external ON event (provider, external_id);

-- overlay: google_event_id -> (provider, external_id)
CREATE TABLE overlay_new (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id        INTEGER NOT NULL REFERENCES period (id) ON DELETE CASCADE,
    provider         TEXT    NOT NULL,
    external_id      TEXT    NOT NULL,
    instance_id      TEXT    NOT NULL DEFAULT '',
    category_id      INTEGER REFERENCES category (id) ON DELETE SET NULL,
    resolved_overlap TEXT,
    note             TEXT    NOT NULL DEFAULT '',
    kind             TEXT    NOT NULL DEFAULT 'category' CHECK (kind IN ('category', 'overlap', 'status')),
    created_at       TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at       TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (period_id, provider, external_id, instance_id, kind)
);

INSERT INTO overlay_new (
    id, period_id, provider, external_id, instance_id, category_id,
    resolved_overlap, note, kind, created_at, updated_at
)
SELECT
    id, period_id, 'google', google_event_id, instance_id, category_id,
    resolved_overlap, note, kind, created_at, updated_at
FROM overlay;

DROP TABLE overlay;
ALTER TABLE overlay_new RENAME TO overlay;

CREATE INDEX idx_overlay_period ON overlay (period_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

CREATE TABLE calendar_old (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    google_calendar_id  TEXT    NOT NULL UNIQUE,
    name                TEXT    NOT NULL,
    is_primary          INTEGER NOT NULL DEFAULT 0 CHECK (is_primary IN (0, 1)),
    selected            INTEGER NOT NULL DEFAULT 0 CHECK (selected IN (0, 1)),
    default_category_id INTEGER REFERENCES category (id) ON DELETE SET NULL,
    created_at          TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

INSERT INTO calendar_old (
    id, google_calendar_id, name, is_primary, selected, default_category_id, created_at
)
SELECT id, external_id, name, is_primary, selected, default_category_id, created_at
FROM calendar
WHERE provider = 'google';

DROP TABLE calendar;
ALTER TABLE calendar_old RENAME TO calendar;

CREATE TABLE event_old (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id          INTEGER NOT NULL REFERENCES period (id) ON DELETE CASCADE,
    calendar_id        INTEGER NOT NULL REFERENCES calendar (id) ON DELETE CASCADE,
    google_event_id    TEXT    NOT NULL,
    instance_id        TEXT    NOT NULL DEFAULT '',
    recurring_event_id TEXT    NOT NULL DEFAULT '',
    ical_uid           TEXT    NOT NULL DEFAULT '',
    title              TEXT    NOT NULL,
    description        TEXT    NOT NULL DEFAULT '',
    location           TEXT    NOT NULL DEFAULT '',
    organizer          TEXT    NOT NULL DEFAULT '',
    attendees          TEXT    NOT NULL DEFAULT '[]',
    status             TEXT    NOT NULL DEFAULT '',
    all_day            INTEGER NOT NULL DEFAULT 0 CHECK (all_day IN (0, 1)),
    start_utc          TEXT,
    end_utc            TEXT,
    start_date         TEXT,
    end_date           TEXT,
    original_tz        TEXT    NOT NULL DEFAULT '',
    active             INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
    source_hash        TEXT    NOT NULL,
    created_at         TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at         TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (period_id, calendar_id, google_event_id, instance_id)
);

INSERT INTO event_old (
    id, period_id, calendar_id, google_event_id, instance_id, recurring_event_id,
    ical_uid, title, description, location, organizer, attendees, status, all_day,
    start_utc, end_utc, start_date, end_date, original_tz, active, source_hash,
    created_at, updated_at
)
SELECT
    id, period_id, calendar_id, external_id, instance_id, recurring_event_id,
    ical_uid, title, description, location, organizer, attendees, status, all_day,
    start_utc, end_utc, start_date, end_date, original_tz, active, source_hash,
    created_at, updated_at
FROM event
WHERE provider = 'google';

DROP TABLE event;
ALTER TABLE event_old RENAME TO event;

CREATE INDEX idx_event_period ON event (period_id);
CREATE INDEX idx_event_ical_uid ON event (ical_uid) WHERE ical_uid <> '';
CREATE INDEX idx_event_gid ON event (google_event_id);

CREATE TABLE overlay_old (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    period_id        INTEGER NOT NULL REFERENCES period (id) ON DELETE CASCADE,
    google_event_id  TEXT    NOT NULL,
    instance_id      TEXT    NOT NULL DEFAULT '',
    category_id      INTEGER REFERENCES category (id) ON DELETE SET NULL,
    resolved_overlap TEXT,
    note             TEXT    NOT NULL DEFAULT '',
    kind             TEXT    NOT NULL DEFAULT 'category' CHECK (kind IN ('category', 'overlap', 'status')),
    created_at       TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    updated_at       TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    UNIQUE (period_id, google_event_id, instance_id, kind)
);

INSERT INTO overlay_old (
    id, period_id, google_event_id, instance_id, category_id,
    resolved_overlap, note, kind, created_at, updated_at
)
SELECT
    id, period_id, external_id, instance_id, category_id,
    resolved_overlap, note, kind, created_at, updated_at
FROM overlay
WHERE provider = 'google';

DROP TABLE overlay;
ALTER TABLE overlay_old RENAME TO overlay;

CREATE INDEX idx_overlay_period ON overlay (period_id);

-- +goose StatementEnd
