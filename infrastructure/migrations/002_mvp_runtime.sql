ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS service_type VARCHAR(32) NOT NULL DEFAULT 'bike',
    ADD COLUMN IF NOT EXISTS last_idle_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL DEFAULT 1;

ALTER TABLE trips
    ADD COLUMN IF NOT EXISTS base_fare_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS distance_fare_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS discount_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS arrived_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS cancellation_reason VARCHAR(160) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL DEFAULT 1;

CREATE INDEX IF NOT EXISTS drivers_availability_service_idx
    ON drivers (approval_status, availability_status, service_type);

CREATE UNIQUE INDEX IF NOT EXISTS one_active_trip_per_driver_idx
    ON trips (driver_id)
    WHERE driver_id IS NOT NULL
      AND status IN ('accepted', 'arriving', 'arrived', 'in_progress');

CREATE TABLE IF NOT EXISTS idempotency_records (
    scope VARCHAR(160) NOT NULL,
    key VARCHAR(200) NOT NULL,
    fingerprint CHAR(64) NOT NULL,
    resource_id VARCHAR(200) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (scope, key)
);

CREATE INDEX IF NOT EXISTS idempotency_expires_idx ON idempotency_records (expires_at);
