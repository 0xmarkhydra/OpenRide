BEGIN;

CREATE TABLE IF NOT EXISTS driver_tariffs (
    id TEXT PRIMARY KEY,
    instance_id TEXT NOT NULL,
    driver_id TEXT NOT NULL,
    service_type TEXT NOT NULL,
    quote_mode TEXT NOT NULL CHECK (quote_mode IN ('manual', 'auto', 'hybrid')),
    currency CHAR(3) NOT NULL,
    base_fare_minor BIGINT NOT NULL DEFAULT 0 CHECK (base_fare_minor >= 0),
    minimum_fare_minor BIGINT NOT NULL DEFAULT 0 CHECK (minimum_fare_minor >= 0),
    per_km_minor BIGINT NOT NULL DEFAULT 0 CHECK (per_km_minor >= 0),
    per_minute_minor BIGINT NOT NULL DEFAULT 0 CHECK (per_minute_minor >= 0),
    pickup_fee_minor BIGINT NOT NULL DEFAULT 0 CHECK (pickup_fee_minor >= 0),
    auto_quote_min_minor BIGINT NOT NULL DEFAULT 0 CHECK (auto_quote_min_minor >= 0),
    auto_quote_max_minor BIGINT NOT NULL DEFAULT 0 CHECK (auto_quote_max_minor >= 0),
    rules JSONB NOT NULL DEFAULT '{}'::jsonb,
    version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0),
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_driver_tariffs_driver_service
    ON driver_tariffs(instance_id, driver_id, service_type)
    WHERE active = TRUE;

CREATE TABLE IF NOT EXISTS mobility_requests (
    id TEXT PRIMARY KEY,
    instance_id TEXT NOT NULL,
    rider_id TEXT NOT NULL,
    service_type TEXT NOT NULL,
    status TEXT NOT NULL,
    pickup_lat DOUBLE PRECISION NOT NULL,
    pickup_lng DOUBLE PRECISION NOT NULL,
    destination_lat DOUBLE PRECISION,
    destination_lng DOUBLE PRECISION,
    attributes JSONB NOT NULL DEFAULT '{}'::jsonb,
    constraints JSONB NOT NULL DEFAULT '{}'::jsonb,
    requested_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ,
    version BIGINT NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_mobility_requests_open
    ON mobility_requests(instance_id, service_type, requested_at DESC)
    WHERE status IN ('open', 'receiving_quotes');

CREATE TABLE IF NOT EXISTS marketplace_quotes (
    id TEXT PRIMARY KEY,
    request_id TEXT NOT NULL REFERENCES mobility_requests(id),
    driver_id TEXT NOT NULL,
    driver_vehicle_id TEXT,
    tariff_id TEXT REFERENCES driver_tariffs(id),
    tariff_version BIGINT,
    status TEXT NOT NULL,
    currency CHAR(3) NOT NULL,
    fare_minor BIGINT NOT NULL CHECK (fare_minor >= 0),
    pickup_distance_m BIGINT NOT NULL DEFAULT 0 CHECK (pickup_distance_m >= 0),
    pickup_eta_s BIGINT NOT NULL DEFAULT 0 CHECK (pickup_eta_s >= 0),
    explanation JSONB NOT NULL DEFAULT '[]'::jsonb,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    accepted_at TIMESTAMPTZ,
    UNIQUE(request_id, driver_id, id)
);

CREATE INDEX IF NOT EXISTS idx_marketplace_quotes_request
    ON marketplace_quotes(request_id, status, fare_minor, pickup_eta_s);

CREATE TABLE IF NOT EXISTS marketplace_agreements (
    id TEXT PRIMARY KEY,
    instance_id TEXT NOT NULL,
    request_id TEXT NOT NULL UNIQUE REFERENCES mobility_requests(id),
    quote_id TEXT NOT NULL UNIQUE REFERENCES marketplace_quotes(id),
    rider_id TEXT NOT NULL,
    driver_id TEXT NOT NULL,
    driver_vehicle_id TEXT,
    service_type TEXT NOT NULL,
    currency CHAR(3) NOT NULL,
    fare_minor BIGINT NOT NULL CHECK (fare_minor >= 0),
    terms_snapshot JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS outbox_events (
    id TEXT PRIMARY KEY,
    event_name TEXT NOT NULL,
    aggregate_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    instance_id TEXT NOT NULL,
    correlation_id TEXT,
    causation_id TEXT,
    payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    published_at TIMESTAMPTZ,
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_outbox_unpublished
    ON outbox_events(created_at)
    WHERE published_at IS NULL;

CREATE TABLE IF NOT EXISTS inbox_messages (
    consumer_name TEXT NOT NULL,
    message_id TEXT NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    processed_at TIMESTAMPTZ,
    PRIMARY KEY (consumer_name, message_id)
);

COMMIT;
