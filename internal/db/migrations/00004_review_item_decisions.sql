-- +goose Up
-- +goose StatementBegin
ALTER TABLE review_item ADD COLUMN conflict_key TEXT NOT NULL DEFAULT '';
ALTER TABLE review_item ADD COLUMN decision_action TEXT NOT NULL DEFAULT '';
ALTER TABLE review_item ADD COLUMN decision_payload TEXT NOT NULL DEFAULT '{}';

CREATE UNIQUE INDEX idx_review_item_period_kind_conflict_key
    ON review_item (period_id, kind, conflict_key)
    WHERE conflict_key <> '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_review_item_period_kind_conflict_key;
-- SQLite cannot drop columns without rebuilding the table; leave added columns in place on down.
-- +goose StatementEnd
