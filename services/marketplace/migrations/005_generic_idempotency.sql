BEGIN;

ALTER TABLE marketplace_idempotency
    ADD COLUMN IF NOT EXISTS request_hash TEXT,
    ADD COLUMN IF NOT EXISTS resource_type TEXT,
    ADD COLUMN IF NOT EXISTS resource_id TEXT;

CREATE INDEX IF NOT EXISTS idx_marketplace_idempotency_resource
    ON marketplace_idempotency(resource_type, resource_id)
    WHERE resource_id IS NOT NULL;

COMMIT;
