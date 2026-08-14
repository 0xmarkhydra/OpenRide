-- FlashX runtime operational settings.
-- Only safe operational knobs live here; secrets/infrastructure stay in env.

CREATE TABLE IF NOT EXISTS operational_settings (
    id TEXT PRIMARY KEY,
    designated_driver_car_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    designated_driver_bike_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    vehicle_inspection_assist_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    dispatch_max_distance_m BIGINT NOT NULL DEFAULT 5000 CHECK (dispatch_max_distance_m BETWEEN 1000 AND 30000),
    driver_location_max_age_seconds INTEGER NOT NULL DEFAULT 20 CHECK (driver_location_max_age_seconds BETWEEN 5 AND 120),
    version BIGINT NOT NULL DEFAULT 1,
    updated_by TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT operational_settings_singleton CHECK (id = 'default')
);

INSERT INTO operational_settings (
    id, designated_driver_car_enabled, designated_driver_bike_enabled,
    vehicle_inspection_assist_enabled, dispatch_max_distance_m,
    driver_location_max_age_seconds, version, updated_at
) VALUES ('default', TRUE, TRUE, TRUE, 5000, 20, 1, NOW())
ON CONFLICT (id) DO NOTHING;
