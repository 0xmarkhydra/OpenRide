BEGIN;

-- Public V2 serializes money minor units as JSON numbers. Keep persisted values
-- inside JavaScript's exact integer range so every supported public client sees
-- the same commercial amount without precision loss.
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'driver_tariffs_base_fare_js_safe') THEN
        ALTER TABLE driver_tariffs ADD CONSTRAINT driver_tariffs_base_fare_js_safe CHECK (base_fare_minor <= 9007199254740991);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'driver_tariffs_minimum_fare_js_safe') THEN
        ALTER TABLE driver_tariffs ADD CONSTRAINT driver_tariffs_minimum_fare_js_safe CHECK (minimum_fare_minor <= 9007199254740991);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'driver_tariffs_per_km_js_safe') THEN
        ALTER TABLE driver_tariffs ADD CONSTRAINT driver_tariffs_per_km_js_safe CHECK (per_km_minor <= 9007199254740991);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'driver_tariffs_per_minute_js_safe') THEN
        ALTER TABLE driver_tariffs ADD CONSTRAINT driver_tariffs_per_minute_js_safe CHECK (per_minute_minor <= 9007199254740991);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'driver_tariffs_pickup_fee_js_safe') THEN
        ALTER TABLE driver_tariffs ADD CONSTRAINT driver_tariffs_pickup_fee_js_safe CHECK (pickup_fee_minor <= 9007199254740991);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'driver_tariffs_auto_min_js_safe') THEN
        ALTER TABLE driver_tariffs ADD CONSTRAINT driver_tariffs_auto_min_js_safe CHECK (auto_quote_min_minor <= 9007199254740991);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'driver_tariffs_auto_max_js_safe') THEN
        ALTER TABLE driver_tariffs ADD CONSTRAINT driver_tariffs_auto_max_js_safe CHECK (auto_quote_max_minor <= 9007199254740991);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'marketplace_quotes_fare_js_safe') THEN
        ALTER TABLE marketplace_quotes ADD CONSTRAINT marketplace_quotes_fare_js_safe CHECK (fare_minor <= 9007199254740991);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'marketplace_agreements_fare_js_safe') THEN
        ALTER TABLE marketplace_agreements ADD CONSTRAINT marketplace_agreements_fare_js_safe CHECK (fare_minor <= 9007199254740991);
    END IF;
END $$;

COMMIT;
