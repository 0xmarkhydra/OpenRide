BEGIN;

-- Serialize quote creation against request state changes so a pending quote
-- cannot be inserted after the request has already been cancelled or agreed.
CREATE OR REPLACE FUNCTION marketplace_guard_quote_request_state()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    request_status TEXT;
BEGIN
    IF NEW.status <> 'pending' THEN
        RETURN NEW;
    END IF;

    SELECT status
      INTO request_status
      FROM mobility_requests
     WHERE id = NEW.request_id
     FOR UPDATE;

    IF request_status IS NULL THEN
        RAISE EXCEPTION 'marketplace request % does not exist', NEW.request_id
            USING ERRCODE = '23503';
    END IF;

    IF request_status NOT IN ('open', 'receiving_quotes') THEN
        RAISE EXCEPTION 'marketplace request % is not accepting quotes', NEW.request_id
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_marketplace_quote_request_state ON marketplace_quotes;
CREATE TRIGGER trg_marketplace_quote_request_state
BEFORE INSERT OR UPDATE OF status, request_id ON marketplace_quotes
FOR EACH ROW
EXECUTE FUNCTION marketplace_guard_quote_request_state();

-- Agreements are commercial records. Corrections should be represented by a
-- new business event/record, never by silently mutating accepted terms.
CREATE OR REPLACE FUNCTION marketplace_reject_agreement_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'marketplace agreements are append-only'
        USING ERRCODE = '55000';
END;
$$;

DROP TRIGGER IF EXISTS trg_marketplace_agreement_append_only ON marketplace_agreements;
CREATE TRIGGER trg_marketplace_agreement_append_only
BEFORE UPDATE OR DELETE ON marketplace_agreements
FOR EACH ROW
EXECUTE FUNCTION marketplace_reject_agreement_mutation();

COMMIT;
