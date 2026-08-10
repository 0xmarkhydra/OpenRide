-- MVP release hardening: one payment ledger record per trip and durable push tokens.
CREATE UNIQUE INDEX IF NOT EXISTS payments_trip_unique_idx ON payments (trip_id);
CREATE INDEX IF NOT EXISTS payments_status_created_idx ON payments (status, created_at DESC);

CREATE TABLE IF NOT EXISTS push_devices (
    id TEXT PRIMARY KEY,
    actor_id TEXT NOT NULL,
    actor_role VARCHAR(16) NOT NULL,
    platform VARCHAR(16) NOT NULL,
    token TEXT NOT NULL UNIQUE,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS push_devices_actor_idx
    ON push_devices (actor_id, actor_role)
    WHERE enabled = TRUE;
