-- Baseline schema: hash-chained audit log and envelope encryption canary (Stage 0).

-- +goose Up
CREATE TABLE IF NOT EXISTS audit_events (
    id INTEGER PRIMARY KEY,
    prev_hash BLOB NOT NULL,
    row_hash BLOB NOT NULL,
    occurred_at TEXT NOT NULL,
    actor_id TEXT NOT NULL DEFAULT '',
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL DEFAULT '',
    resource_id TEXT NOT NULL DEFAULT '',
    request_id TEXT NOT NULL DEFAULT '',
    before_hash TEXT NOT NULL DEFAULT '',
    after_hash TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_audit_events_occurred_at ON audit_events (occurred_at);

CREATE TABLE IF NOT EXISTS audit_integrity_events (
    id INTEGER PRIMARY KEY,
    detected_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
    break_at_event_id INTEGER,
    detail TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS envelope_canary (
    id INTEGER PRIMARY KEY,
    key_id TEXT NOT NULL,
    wrapped_dek BLOB NOT NULL,
    ciphertext BLOB NOT NULL,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

-- +goose Down
DROP TABLE IF EXISTS envelope_canary;
DROP TABLE IF EXISTS audit_integrity_events;
DROP TABLE IF EXISTS audit_events;
