-- OpenRide Marketplace V2 foundation.
-- Additive only: legacy trips/pricing/dispatch remain operational during migration.

CREATE TABLE IF NOT EXISTS instances (
    id TEXT PRIMARY KEY,
    slug VARCHAR(80) NOT NULL UNIQUE,
    name VARCHAR(160) NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'suspended', 'disabled')),
    currency CHAR(3) NOT NULL DEFAULT 'VND',
    country_code CHAR(2) NOT NULL DEFAULT 'VN',
    timezone VARCHAR(80) NOT NULL DEFAULT 'Asia/Ho_Chi_Minh',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO instances (id, slug, name)
VALUES ('instance_default', 'default', 'OpenRide Default')
ON CONFLICT (id) DO NOTHING;

CREATE TABLE IF NOT EXISTS driver_tariffs (
    id TEXT PRIMARY KEY,
    instance_id TEXT NOT NULL REFERENCES instances(id),
    driver_id TEXT NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    service_type VARCHAR(48) NOT NULL,
    quote_mode VARCHAR(16) NOT NULL DEFAULT 'manual'
        CHECK (quote_mode IN ('manual', 'auto', 'hybrid')),
    currency CHAR(3) NOT NULL DEFAULT 'VND',
    base_fare_minor BIGINT NOT NULL DEFAULT 0 CHECK (base_fare_minor >= 0),
    minimum_fare_minor BIGINT NOT NULL DEFAULT 0 CHECK (minimum_fare_minor >= 0),
    per_km_minor BIGINT NOT NULL DEFAULT 0 CHECK (per_km_minor >= 0),
    per_minute_minor BIGINT NOT NULL DEFAULT 0 CHECK (per_minute_minor >= 0),
    pickup_fee_minor BIGINT NOT NULL DEFAULT 0 CHECK (pickup_fee_minor >= 0),
    auto_quote_min_minor BIGINT NOT NULL DEFAULT 0 CHECK (auto_quote_min_minor >= 0),
    auto_quote_max_minor BIGINT NOT NULL DEFAULT 0 CHECK (auto_quote_max_minor >= 0),
    status VARCHAR(24) NOT NULL DEFAULT 'active'
        CHECK (status IN ('draft', 'active', 'paused', 'archived')),
    version BIGINT NOT NULL DEFAULT 1,
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_to TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (effective_to IS NULL OR effective_to > effective_from),
    CHECK (auto_quote_max_minor = 0 OR auto_quote_max_minor >= auto_quote_min_minor)
);

CREATE INDEX IF NOT EXISTS driver_tariffs_driver_service_idx
    ON driver_tariffs (driver_id, service_type, status, created_at DESC);
CREATE INDEX IF NOT EXISTS driver_tariffs_instance_service_idx
    ON driver_tariffs (instance_id, service_type, status, created_at DESC);

CREATE TABLE IF NOT EXISTS driver_tariff_rules (
    id TEXT PRIMARY KEY,
    tariff_id TEXT NOT NULL REFERENCES driver_tariffs(id) ON DELETE CASCADE,
    rule_type VARCHAR(48) NOT NULL,
    priority INTEGER NOT NULL DEFAULT 100,
    conditions JSONB NOT NULL DEFAULT '{}'::jsonb,
    adjustment_type VARCHAR(24) NOT NULL
        CHECK (adjustment_type IN ('fixed', 'percent', 'per_km', 'per_minute', 'override')),
    adjustment_value BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(24) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active', 'paused', 'archived')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS driver_tariff_rules_tariff_idx
    ON driver_tariff_rules (tariff_id, status, priority, created_at);

CREATE TABLE IF NOT EXISTS mobility_requests (
    id TEXT PRIMARY KEY,
    instance_id TEXT NOT NULL REFERENCES instances(id),
    rider_id TEXT NOT NULL REFERENCES users(id),
    service_type VARCHAR(48) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'open', 'receiving_quotes', 'agreed', 'closed', 'cancelled', 'expired')),
    pickup_lat DOUBLE PRECISION NOT NULL,
    pickup_lng DOUBLE PRECISION NOT NULL,
    destination_lat DOUBLE PRECISION NOT NULL,
    destination_lng DOUBLE PRECISION NOT NULL,
    estimated_distance_m BIGINT NOT NULL DEFAULT 0 CHECK (estimated_distance_m >= 0),
    estimated_duration_s BIGINT NOT NULL DEFAULT 0 CHECK (estimated_duration_s >= 0),
    preferences JSONB NOT NULL DEFAULT '{}'::jsonb,
    constraints JSONB NOT NULL DEFAULT '{}'::jsonb,
    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    agreed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    cancellation_reason TEXT NOT NULL DEFAULT '',
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (pickup_lat BETWEEN -90 AND 90),
    CHECK (pickup_lng BETWEEN -180 AND 180),
    CHECK (destination_lat BETWEEN -90 AND 90),
    CHECK (destination_lng BETWEEN -180 AND 180),
    CHECK (expires_at IS NULL OR expires_at > requested_at)
);

CREATE INDEX IF NOT EXISTS mobility_requests_instance_status_idx
    ON mobility_requests (instance_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS mobility_requests_rider_created_idx
    ON mobility_requests (rider_id, created_at DESC);
CREATE INDEX IF NOT EXISTS mobility_requests_expiry_idx
    ON mobility_requests (expires_at)
    WHERE status IN ('open', 'receiving_quotes') AND expires_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS marketplace_quotes (
    id TEXT PRIMARY KEY,
    request_id TEXT NOT NULL REFERENCES mobility_requests(id) ON DELETE CASCADE,
    driver_id TEXT NOT NULL REFERENCES drivers(id) ON DELETE CASCADE,
    driver_vehicle_id TEXT REFERENCES vehicles(id),
    tariff_id TEXT REFERENCES driver_tariffs(id),
    tariff_version BIGINT,
    status VARCHAR(24) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'rejected', 'withdrawn', 'expired', 'invalidated')),
    fare_total_minor BIGINT NOT NULL CHECK (fare_total_minor >= 0),
    currency CHAR(3) NOT NULL DEFAULT 'VND',
    pricing_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    pickup_distance_m BIGINT NOT NULL DEFAULT 0 CHECK (pickup_distance_m >= 0),
    pickup_eta_s BIGINT NOT NULL DEFAULT 0 CHECK (pickup_eta_s >= 0),
    ranking_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    withdrawn_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS marketplace_quotes_request_status_idx
    ON marketplace_quotes (request_id, status, created_at);
CREATE INDEX IF NOT EXISTS marketplace_quotes_driver_status_idx
    ON marketplace_quotes (driver_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS marketplace_quotes_expiry_idx
    ON marketplace_quotes (expires_at)
    WHERE status = 'pending';
CREATE UNIQUE INDEX IF NOT EXISTS marketplace_quotes_request_driver_pending_idx
    ON marketplace_quotes (request_id, driver_id)
    WHERE status = 'pending';

CREATE TABLE IF NOT EXISTS marketplace_agreements (
    id TEXT PRIMARY KEY,
    instance_id TEXT NOT NULL REFERENCES instances(id),
    request_id TEXT NOT NULL UNIQUE REFERENCES mobility_requests(id),
    quote_id TEXT NOT NULL UNIQUE REFERENCES marketplace_quotes(id),
    rider_id TEXT NOT NULL REFERENCES users(id),
    driver_id TEXT NOT NULL REFERENCES drivers(id),
    driver_vehicle_id TEXT REFERENCES vehicles(id),
    service_type VARCHAR(48) NOT NULL,
    pickup_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    destination_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    route_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    fare_total_minor BIGINT NOT NULL CHECK (fare_total_minor >= 0),
    currency CHAR(3) NOT NULL DEFAULT 'VND',
    pricing_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    terms_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    policy_version VARCHAR(80) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS marketplace_agreements_driver_created_idx
    ON marketplace_agreements (driver_id, created_at DESC);
CREATE INDEX IF NOT EXISTS marketplace_agreements_rider_created_idx
    ON marketplace_agreements (rider_id, created_at DESC);

CREATE TABLE IF NOT EXISTS rides_v2 (
    id TEXT PRIMARY KEY,
    agreement_id TEXT NOT NULL UNIQUE REFERENCES marketplace_agreements(id) ON DELETE RESTRICT,
    instance_id TEXT NOT NULL REFERENCES instances(id),
    rider_id TEXT NOT NULL REFERENCES users(id),
    driver_id TEXT NOT NULL REFERENCES drivers(id),
    service_type VARCHAR(48) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'assigned'
        CHECK (status IN ('assigned', 'driver_en_route', 'driver_arrived', 'passenger_onboard', 'in_progress', 'completed', 'cancelled', 'failed')),
    version BIGINT NOT NULL DEFAULT 1,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    arrived_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    cancellation_reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS rides_v2_driver_status_idx
    ON rides_v2 (driver_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS rides_v2_rider_created_idx
    ON rides_v2 (rider_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS rides_v2_one_active_per_driver_idx
    ON rides_v2 (driver_id)
    WHERE status NOT IN ('completed', 'cancelled', 'failed');

CREATE TABLE IF NOT EXISTS ride_v2_status_history (
    id BIGSERIAL PRIMARY KEY,
    ride_id TEXT NOT NULL REFERENCES rides_v2(id) ON DELETE CASCADE,
    sequence BIGINT NOT NULL,
    status VARCHAR(32) NOT NULL,
    actor_type VARCHAR(32) NOT NULL,
    actor_id TEXT,
    reason_code VARCHAR(64) NOT NULL DEFAULT '',
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (ride_id, sequence)
);

-- Explicitly keep legacy pricing_rules, trips and dispatch_offers untouched.
-- They remain the compatibility path until V2 request/quote/agreement/ride APIs are implemented.
