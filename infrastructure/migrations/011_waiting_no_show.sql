-- FlashX pickup waiting/no-show policy.
-- The pilot baseline is 10 minutes, but Operations must be able to change it
-- without rebuilding either mobile application.

ALTER TABLE operational_settings
    ADD COLUMN IF NOT EXISTS pickup_grace_period_seconds INTEGER NOT NULL DEFAULT 600;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'operational_settings_pickup_grace_period_check'
    ) THEN
        ALTER TABLE operational_settings
            ADD CONSTRAINT operational_settings_pickup_grace_period_check
            CHECK (pickup_grace_period_seconds BETWEEN 60 AND 3600);
    END IF;
END $$;
