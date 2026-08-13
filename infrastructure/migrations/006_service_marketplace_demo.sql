-- FlashX MVP service-marketplace foundation.
-- Additive migration: existing trip/auth/payment foundations stay compatible.

CREATE TABLE IF NOT EXISTS customer_vehicles (
    id TEXT PRIMARY KEY,
    owner_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(16) NOT NULL CHECK (type IN ('car', 'motorbike')),
    license_plate VARCHAR(32) NOT NULL,
    brand VARCHAR(80) NOT NULL DEFAULT '',
    model VARCHAR(80) NOT NULL DEFAULT '',
    year SMALLINT,
    color VARCHAR(40) NOT NULL DEFAULT '',
    transmission VARCHAR(16) NOT NULL DEFAULT 'n/a' CHECK (transmission IN ('automatic', 'manual', 'n/a')),
    seats SMALLINT,
    notes TEXT NOT NULL DEFAULT '',
    photo_object_key TEXT NOT NULL DEFAULT '',
    status VARCHAR(24) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(owner_user_id, license_plate)
);

CREATE INDEX IF NOT EXISTS customer_vehicles_owner_idx
    ON customer_vehicles (owner_user_id, status, created_at DESC);

ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS capabilities TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[];

-- Existing car/bike profiles keep their old primary service_type for compatibility.
UPDATE drivers
SET capabilities = CASE service_type
    WHEN 'car' THEN ARRAY['designated_driver_car']::TEXT[]
    WHEN 'bike' THEN ARRAY['designated_driver_bike']::TEXT[]
    WHEN 'designated_driver_car' THEN ARRAY['designated_driver_car']::TEXT[]
    WHEN 'designated_driver_bike' THEN ARRAY['designated_driver_bike']::TEXT[]
    WHEN 'vehicle_inspection_assist' THEN ARRAY['vehicle_inspection_assist','designated_driver_car']::TEXT[]
    ELSE capabilities
END
WHERE cardinality(capabilities) = 0;

ALTER TABLE trips
    ADD COLUMN IF NOT EXISTS customer_vehicle_id TEXT REFERENCES customer_vehicles(id),
    ADD COLUMN IF NOT EXISTS booking_mode VARCHAR(16) NOT NULL DEFAULT 'immediate',
    ADD COLUMN IF NOT EXISTS scheduled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS service_fare_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS schedule_fare_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS surcharge_minor BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS inspection_result VARCHAR(24) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS incident_type VARCHAR(48) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS incident_note TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS incident_open BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS vehicle_received_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS handover_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS trips_scheduled_idx
    ON trips (scheduled_at)
    WHERE status = 'scheduled';
CREATE INDEX IF NOT EXISTS trips_customer_vehicle_idx
    ON trips (customer_vehicle_id, status)
    WHERE customer_vehicle_id IS NOT NULL;

DROP INDEX IF EXISTS one_active_trip_per_driver_idx;
CREATE UNIQUE INDEX one_active_trip_per_driver_idx
    ON trips (driver_id)
    WHERE driver_id IS NOT NULL
      AND status NOT IN ('scheduled', 'searching', 'completed', 'cancelled');

INSERT INTO pricing_rules (id, service_type, base_fare_minor, per_km_minor, per_minute_minor, minimum_fare_minor)
VALUES
    ('pricing_designated_bike_demo_v1', 'designated_driver_bike', 45000, 9000, 0, 60000),
    ('pricing_designated_car_demo_v1', 'designated_driver_car', 90000, 18000, 0, 120000),
    ('pricing_inspection_demo_v1', 'vehicle_inspection_assist', 299000, 5000, 0, 299000)
ON CONFLICT (id) DO NOTHING;
