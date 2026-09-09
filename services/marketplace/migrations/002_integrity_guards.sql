BEGIN;

-- Keep migrations append-only. This migration tightens invariants introduced
-- by 001 without rewriting an already-published migration.

ALTER TABLE driver_tariffs
    ADD CONSTRAINT chk_driver_tariffs_currency
        CHECK (currency ~ '^[A-Z]{3}$') NOT VALID,
    ADD CONSTRAINT chk_driver_tariffs_auto_quote_bounds
        CHECK (auto_quote_max_minor = 0 OR auto_quote_max_minor >= auto_quote_min_minor) NOT VALID;

ALTER TABLE driver_tariffs VALIDATE CONSTRAINT chk_driver_tariffs_currency;
ALTER TABLE driver_tariffs VALIDATE CONSTRAINT chk_driver_tariffs_auto_quote_bounds;

ALTER TABLE mobility_requests
    ADD CONSTRAINT chk_mobility_requests_status
        CHECK (status IN ('draft', 'open', 'receiving_quotes', 'agreed', 'closed', 'cancelled', 'expired')) NOT VALID,
    ADD CONSTRAINT chk_mobility_requests_pickup_lat
        CHECK (pickup_lat BETWEEN -90 AND 90) NOT VALID,
    ADD CONSTRAINT chk_mobility_requests_pickup_lng
        CHECK (pickup_lng BETWEEN -180 AND 180) NOT VALID,
    ADD CONSTRAINT chk_mobility_requests_destination_pair
        CHECK ((destination_lat IS NULL) = (destination_lng IS NULL)) NOT VALID,
    ADD CONSTRAINT chk_mobility_requests_destination_lat
        CHECK (destination_lat IS NULL OR destination_lat BETWEEN -90 AND 90) NOT VALID,
    ADD CONSTRAINT chk_mobility_requests_destination_lng
        CHECK (destination_lng IS NULL OR destination_lng BETWEEN -180 AND 180) NOT VALID,
    ADD CONSTRAINT chk_mobility_requests_expiry
        CHECK (expires_at IS NULL OR expires_at > requested_at) NOT VALID;

ALTER TABLE mobility_requests VALIDATE CONSTRAINT chk_mobility_requests_status;
ALTER TABLE mobility_requests VALIDATE CONSTRAINT chk_mobility_requests_pickup_lat;
ALTER TABLE mobility_requests VALIDATE CONSTRAINT chk_mobility_requests_pickup_lng;
ALTER TABLE mobility_requests VALIDATE CONSTRAINT chk_mobility_requests_destination_pair;
ALTER TABLE mobility_requests VALIDATE CONSTRAINT chk_mobility_requests_destination_lat;
ALTER TABLE mobility_requests VALIDATE CONSTRAINT chk_mobility_requests_destination_lng;
ALTER TABLE mobility_requests VALIDATE CONSTRAINT chk_mobility_requests_expiry;

-- The old UNIQUE(request_id, driver_id, id) from 001 is redundant because id is
-- already the primary key. Replace it with a composite identity used by the
-- Agreement FK to prove the accepted quote belongs to the same request/driver.
ALTER TABLE marketplace_quotes
    DROP CONSTRAINT IF EXISTS marketplace_quotes_request_id_driver_id_id_key;

ALTER TABLE marketplace_quotes
    ADD CONSTRAINT uq_marketplace_quotes_identity UNIQUE (id, request_id, driver_id),
    ADD CONSTRAINT chk_marketplace_quotes_status
        CHECK (status IN ('pending', 'accepted', 'rejected', 'withdrawn', 'expired', 'invalidated')) NOT VALID,
    ADD CONSTRAINT chk_marketplace_quotes_currency
        CHECK (currency ~ '^[A-Z]{3}$') NOT VALID,
    ADD CONSTRAINT chk_marketplace_quotes_tariff_version
        CHECK (tariff_version IS NULL OR tariff_version > 0) NOT VALID,
    ADD CONSTRAINT chk_marketplace_quotes_expiry
        CHECK (expires_at > created_at) NOT VALID,
    ADD CONSTRAINT chk_marketplace_quotes_accepted_at
        CHECK (
            (status = 'accepted' AND accepted_at IS NOT NULL)
            OR (status <> 'accepted' AND accepted_at IS NULL)
        ) NOT VALID;

ALTER TABLE marketplace_quotes VALIDATE CONSTRAINT chk_marketplace_quotes_status;
ALTER TABLE marketplace_quotes VALIDATE CONSTRAINT chk_marketplace_quotes_currency;
ALTER TABLE marketplace_quotes VALIDATE CONSTRAINT chk_marketplace_quotes_tariff_version;
ALTER TABLE marketplace_quotes VALIDATE CONSTRAINT chk_marketplace_quotes_expiry;
ALTER TABLE marketplace_quotes VALIDATE CONSTRAINT chk_marketplace_quotes_accepted_at;

ALTER TABLE mobility_requests
    ADD CONSTRAINT uq_mobility_requests_identity
        UNIQUE (id, instance_id, rider_id, service_type);

ALTER TABLE marketplace_agreements
    ADD CONSTRAINT chk_marketplace_agreements_currency
        CHECK (currency ~ '^[A-Z]{3}$') NOT VALID,
    ADD CONSTRAINT fk_marketplace_agreements_request_identity
        FOREIGN KEY (request_id, instance_id, rider_id, service_type)
        REFERENCES mobility_requests(id, instance_id, rider_id, service_type),
    ADD CONSTRAINT fk_marketplace_agreements_quote_identity
        FOREIGN KEY (quote_id, request_id, driver_id)
        REFERENCES marketplace_quotes(id, request_id, driver_id);

ALTER TABLE marketplace_agreements VALIDATE CONSTRAINT chk_marketplace_agreements_currency;

COMMIT;
