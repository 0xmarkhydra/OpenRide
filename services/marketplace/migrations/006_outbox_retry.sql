BEGIN;

ALTER TABLE outbox_events
    ADD COLUMN IF NOT EXISTS next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS lease_token TEXT,
    ADD COLUMN IF NOT EXISTS dead_lettered_at TIMESTAMPTZ;

DROP INDEX IF EXISTS idx_outbox_unpublished;

CREATE INDEX IF NOT EXISTS idx_outbox_ready
    ON outbox_events(next_attempt_at, created_at)
    WHERE published_at IS NULL AND dead_lettered_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_outbox_dead_lettered
    ON outbox_events(dead_lettered_at)
    WHERE dead_lettered_at IS NOT NULL;

COMMIT;
