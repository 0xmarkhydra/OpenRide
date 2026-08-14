-- FlashX Admin Operations v2: durable, versioned pricing rules.
-- Pricing updates are append-only versions: one active rule per service type.

ALTER TABLE pricing_rules
    ADD COLUMN IF NOT EXISTS service_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS pricing_version TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS updated_by TEXT,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE pricing_rules
SET pricing_version = id
WHERE pricing_version = '';

-- Preserve the current inspection fare breakdown used by the API before
-- pricing became database-backed: 299k is a service fee, not a base fare.
UPDATE pricing_rules
SET service_minor = base_fare_minor,
    base_fare_minor = 0,
    pricing_version = COALESCE(NULLIF(pricing_version, ''), id),
    updated_at = NOW()
WHERE service_type = 'vehicle_inspection_assist'
  AND id = 'pricing_inspection_demo_v1'
  AND service_minor = 0
  AND base_fare_minor = 299000;

-- Keep only the newest active row if a previous deployment accidentally left
-- multiple active versions for the same exact service_type.
WITH ranked AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY service_type ORDER BY effective_at DESC, created_at DESC, id DESC) AS rn
    FROM pricing_rules
    WHERE active = TRUE
)
UPDATE pricing_rules p
SET active = FALSE, updated_at = NOW()
FROM ranked r
WHERE p.id = r.id AND r.rn > 1;

CREATE UNIQUE INDEX IF NOT EXISTS pricing_rules_one_active_service_idx
    ON pricing_rules (service_type)
    WHERE active = TRUE;
