CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    phone VARCHAR(32) UNIQUE NOT NULL,
    full_name VARCHAR(160) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS drivers (
    id TEXT PRIMARY KEY,
    phone VARCHAR(32) UNIQUE,
    full_name VARCHAR(160) NOT NULL DEFAULT '',
    service_type VARCHAR(32) NOT NULL DEFAULT 'bike',
    approval_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    availability_status VARCHAR(32) NOT NULL DEFAULT 'offline',
    latest_location GEOGRAPHY(POINT, 4326),
    location_accuracy_m DOUBLE PRECISION,
    heading_deg DOUBLE PRECISION,
    speed_mps DOUBLE PRECISION,
    location_captured_at TIMESTAMPTZ,
    last_idle_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS drivers_location_gix ON drivers USING GIST (latest_location);
CREATE INDEX IF NOT EXISTS drivers_dispatch_idx ON drivers (service_type, approval_status, availability_status);

CREATE TABLE IF NOT EXISTS vehicles (
    id TEXT PRIMARY KEY,
    driver_id TEXT NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    service_type VARCHAR(32) NOT NULL,
    plate_number VARCHAR(32) NOT NULL,
    brand VARCHAR(80) NOT NULL DEFAULT '',
    model VARCHAR(80) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS driver_documents (
    id TEXT PRIMARY KEY,
    driver_id TEXT NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    document_type VARCHAR(48) NOT NULL,
    object_key TEXT NOT NULL,
    review_status VARCHAR(32) NOT NULL DEFAULT 'pending',
    review_note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS trips (
    id TEXT PRIMARY KEY,
    rider_id TEXT NOT NULL,
    driver_id TEXT REFERENCES drivers(id),
    service_type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'searching',
    pickup GEOGRAPHY(POINT, 4326) NOT NULL,
    destination GEOGRAPHY(POINT, 4326) NOT NULL,
    estimated_distance_m BIGINT NOT NULL DEFAULT 0,
    estimated_duration_s BIGINT NOT NULL DEFAULT 0,
    estimated_fare_minor BIGINT NOT NULL DEFAULT 0,
    final_fare_minor BIGINT NOT NULL DEFAULT 0,
    base_fare_minor BIGINT NOT NULL DEFAULT 0,
    distance_fare_minor BIGINT NOT NULL DEFAULT 0,
    discount_minor BIGINT NOT NULL DEFAULT 0,
    currency CHAR(3) NOT NULL DEFAULT 'VND',
    cancellation_reason TEXT NOT NULL DEFAULT '',
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    accepted_at TIMESTAMPTZ,
    arrived_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS trips_pickup_gix ON trips USING GIST (pickup);
CREATE INDEX IF NOT EXISTS trips_destination_gix ON trips USING GIST (destination);
CREATE INDEX IF NOT EXISTS trips_status_created_idx ON trips (status, created_at DESC);
CREATE INDEX IF NOT EXISTS trips_rider_created_idx ON trips (rider_id, created_at DESC);
CREATE INDEX IF NOT EXISTS trips_driver_status_idx ON trips (driver_id, status) WHERE driver_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS trip_status_history (
    id BIGSERIAL PRIMARY KEY,
    trip_id TEXT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL,
    actor_type VARCHAR(32) NOT NULL,
    actor_id TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS dispatch_offers (
    id TEXT PRIMARY KEY,
    trip_id TEXT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    driver_id TEXT NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    distance_to_pickup_m BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS dispatch_offers_driver_status_idx ON dispatch_offers (driver_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS dispatch_offers_trip_idx ON dispatch_offers (trip_id, created_at DESC);

CREATE TABLE IF NOT EXISTS idempotency_keys (
    scope TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours',
    PRIMARY KEY (scope, idempotency_key)
);

CREATE TABLE IF NOT EXISTS pricing_rules (
    id TEXT PRIMARY KEY,
    service_type VARCHAR(32) NOT NULL,
    base_fare_minor BIGINT NOT NULL,
    per_km_minor BIGINT NOT NULL,
    per_minute_minor BIGINT NOT NULL DEFAULT 0,
    minimum_fare_minor BIGINT NOT NULL DEFAULT 0,
    currency CHAR(3) NOT NULL DEFAULT 'VND',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    effective_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY,
    trip_id TEXT NOT NULL REFERENCES trips(id),
    provider VARCHAR(32) NOT NULL,
    method VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL,
    amount_minor BIGINT NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'VND',
    provider_reference TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ratings (
    id TEXT PRIMARY KEY,
    trip_id TEXT NOT NULL UNIQUE REFERENCES trips(id) ON DELETE CASCADE,
    rider_id TEXT NOT NULL,
    driver_id TEXT NOT NULL REFERENCES drivers(id),
    stars SMALLINT NOT NULL CHECK (stars BETWEEN 1 AND 5),
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS otp_challenges (
    id TEXT PRIMARY KEY,
    phone VARCHAR(32) NOT NULL,
    role VARCHAR(16) NOT NULL,
    code_hash TEXT NOT NULL,
    attempts SMALLINT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_sessions (
    id TEXT PRIMARY KEY,
    actor_id TEXT NOT NULL,
    role VARCHAR(16) NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS admin_users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    role VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    actor_type VARCHAR(32) NOT NULL,
    actor_id TEXT,
    action TEXT NOT NULL,
    resource_type VARCHAR(64) NOT NULL,
    resource_id TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO pricing_rules (id, service_type, base_fare_minor, per_km_minor, minimum_fare_minor)
VALUES
    ('pricing_bike_default', 'bike', 12000, 4500, 15000),
    ('pricing_car_default', 'car', 25000, 11000, 35000)
ON CONFLICT (id) DO NOTHING;
