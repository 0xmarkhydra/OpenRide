-- FlashX custody evidence for customer-owned vehicle receive/return handover.
-- Evidence is separate from trip lifecycle so photos/confirmations can evolve
-- without expanding the state-machine columns on trips.

CREATE TABLE IF NOT EXISTS custody_evidence (
    id TEXT PRIMARY KEY,
    trip_id TEXT NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    stage VARCHAR(16) NOT NULL CHECK (stage IN ('pickup', 'return')),
    condition_note TEXT NOT NULL DEFAULT '',
    odometer_km BIGINT CHECK (odometer_km IS NULL OR odometer_km >= 0),
    fuel_percent SMALLINT CHECK (fuel_percent IS NULL OR fuel_percent BETWEEN 0 AND 100),
    battery_percent SMALLINT CHECK (battery_percent IS NULL OR battery_percent BETWEEN 0 AND 100),
    driver_confirmed_at TIMESTAMPTZ,
    rider_confirmed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (trip_id, stage)
);

CREATE INDEX IF NOT EXISTS custody_evidence_trip_idx
    ON custody_evidence (trip_id, stage);

CREATE TABLE IF NOT EXISTS custody_evidence_photos (
    id TEXT PRIMARY KEY,
    evidence_id TEXT NOT NULL REFERENCES custody_evidence(id) ON DELETE CASCADE,
    photo_type VARCHAR(24) NOT NULL,
    object_key TEXT NOT NULL UNIQUE,
    filename TEXT NOT NULL,
    content_type VARCHAR(64) NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS custody_evidence_photos_evidence_idx
    ON custody_evidence_photos (evidence_id, created_at ASC);
