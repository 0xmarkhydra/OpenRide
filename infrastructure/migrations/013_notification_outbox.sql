-- Durable push notification outbox. Existing push_devices table was introduced in migration 004.
CREATE TABLE IF NOT EXISTS notification_outbox (
    id TEXT PRIMARY KEY,
    actor_id TEXT NOT NULL,
    actor_role VARCHAR(16) NOT NULL,
    event_type VARCHAR(96) NOT NULL,
    resource_id TEXT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT notification_outbox_role_chk CHECK (actor_role IN ('rider','driver')),
    CONSTRAINT notification_outbox_status_chk CHECK (status IN ('pending','sent','skipped','failed'))
);

CREATE INDEX IF NOT EXISTS notification_outbox_pending_idx
    ON notification_outbox (created_at ASC)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS notification_outbox_actor_idx
    ON notification_outbox (actor_id, actor_role, created_at DESC);
