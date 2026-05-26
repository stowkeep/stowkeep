-- PostgreSQL baseline schema (portable equivalent of shared/00001_baseline.sql).

-- +goose Up
CREATE TABLE IF NOT EXISTS audit_events (
    id BIGSERIAL PRIMARY KEY,
    prev_hash BYTEA NOT NULL,
    row_hash BYTEA NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    actor_id TEXT NOT NULL DEFAULT '',
    action TEXT NOT NULL,
    resource_type TEXT NOT NULL DEFAULT '',
    resource_id TEXT NOT NULL DEFAULT '',
    request_id TEXT NOT NULL DEFAULT '',
    before_hash TEXT NOT NULL DEFAULT '',
    after_hash TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_events_occurred_at ON audit_events (occurred_at);

CREATE TABLE IF NOT EXISTS audit_integrity_events (
    id BIGSERIAL PRIMARY KEY,
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    break_at_event_id BIGINT,
    detail TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS envelope_canary (
    id BIGSERIAL PRIMARY KEY,
    key_id TEXT NOT NULL,
    wrapped_dek BYTEA NOT NULL,
    ciphertext BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS envelope_canary;
DROP TABLE IF EXISTS audit_integrity_events;
DROP TABLE IF EXISTS audit_events;
