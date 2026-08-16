-- Driver qualification metadata used by dispatch safety rules.
-- Existing approved drivers keep their marketplace capabilities from migration 006;
-- manual-transmission eligibility is deliberately opt-in.

ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS license_class VARCHAR(32) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS license_expiry DATE,
    ADD COLUMN IF NOT EXISTS can_drive_manual BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS drivers_qualification_idx
    ON drivers (approval_status, availability_status, can_drive_manual);
