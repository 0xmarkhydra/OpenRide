BEGIN;

CREATE TABLE IF NOT EXISTS marketplace_idempotency (
    actor_id TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    command_name TEXT NOT NULL,
    agreement_id TEXT REFERENCES marketplace_agreements(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (actor_id, idempotency_key, command_name)
);

CREATE INDEX IF NOT EXISTS idx_marketplace_idempotency_agreement
    ON marketplace_idempotency(agreement_id)
    WHERE agreement_id IS NOT NULL;

COMMIT;
