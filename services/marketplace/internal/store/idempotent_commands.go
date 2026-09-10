package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
)

var ErrIdempotencyConflict = errors.New("marketplace store: idempotency key reused with different payload")

type idempotencyRecord struct {
	RequestHash   string
	ResourceType  string
	ResourceID    string
	ResponseStatus int
	ResponseBody  []byte
}

func lockCommandIdempotency(ctx context.Context, tx pgx.Tx, actorID, key, command string) error {
	_, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, actorID+":"+key+":"+command)
	return err
}

func findCommandIdempotency(ctx context.Context, tx pgx.Tx, actorID, key, command string) (idempotencyRecord, bool, error) {
	var rec idempotencyRecord
	err := tx.QueryRow(ctx, `SELECT COALESCE(request_hash,''),COALESCE(resource_type,''),COALESCE(resource_id,''),COALESCE(response_status,0),COALESCE(response_body,'null'::jsonb)
		FROM marketplace_idempotency WHERE actor_id=$1 AND idempotency_key=$2 AND command_name=$3`, actorID, key, command).
		Scan(&rec.RequestHash, &rec.ResourceType, &rec.ResourceID, &rec.ResponseStatus, &rec.ResponseBody)
	if errors.Is(err, pgx.ErrNoRows) {
		return idempotencyRecord{}, false, nil
	}
	return rec, err == nil, err
}

func saveCommandIdempotency(ctx context.Context, tx pgx.Tx, actorID, key, command, requestHash, resourceType, resourceID string, responseStatus int, response any) error {
	body, err := json.Marshal(response)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO marketplace_idempotency(actor_id,idempotency_key,command_name,request_hash,resource_type,resource_id,response_status,response_body)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, actorID, key, command, requestHash, resourceType, resourceID, responseStatus, body)
	return err
}

func verifyReplay(rec idempotencyRecord, requestHash, resourceType string) error {
	if rec.RequestHash != requestHash || rec.ResourceType != resourceType || rec.ResourceID == "" || rec.ResponseStatus == 0 || len(rec.ResponseBody) == 0 {
		return ErrIdempotencyConflict
	}
	return nil
}

func decodeReplay(rec idempotencyRecord, requestHash, resourceType string, dst any) error {
	if err := verifyReplay(rec, requestHash, resourceType); err != nil {
		return err
	}
	return json.Unmarshal(rec.ResponseBody, dst)
}

func (s *Store) withinSerializable(ctx context.Context, fn func(pgx.Tx) error) error {
	var lastErr error
	for attempt := 0; attempt < serializableTxMaxAttempts; attempt++ {
		tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
		if err != nil {
			return err
		}
		err = fn(tx)
		if err == nil {
			err = tx.Commit(ctx)
		} else {
			_ = tx.Rollback(context.Background())
		}
		if err == nil {
			return nil
		}
		lastErr = err
		if !isRetryableTxError(err) {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
	}
	return lastErr
}

func (s *Store) CreateRequestIdempotent(ctx context.Context, actorID, key, requestHash string, r marketplace.Request) (marketplace.Request, bool, error) {
	if err := r.Validate(); err != nil {
		return marketplace.Request{}, false, err
	}
	var result marketplace.Request
	var replay bool
	err := s.withinSerializable(ctx, func(tx pgx.Tx) error {
		if err := lockCommandIdempotency(ctx, tx, actorID, key, "create_request"); err != nil {
			return err
		}
		rec, found, err := findCommandIdempotency(ctx, tx, actorID, key, "create_request")
		if err != nil {
			return err
		}
		if found {
			if err := decodeReplay(rec, requestHash, "request", &result); err != nil {
				return err
			}
			replay = true
			return nil
		}
		attrs, err := json.Marshal(r.Attributes)
		if err != nil {
			return err
		}
		constraints, err := json.Marshal(r.Constraints)
		if err != nil {
			return err
		}
		var dlat, dlng any
		if r.Destination != nil {
			dlat, dlng = r.Destination.Lat, r.Destination.Lng
		}
		_, err = tx.Exec(ctx, `INSERT INTO mobility_requests
			(id,instance_id,rider_id,service_type,status,pickup_lat,pickup_lng,destination_lat,destination_lng,attributes,constraints,requested_at,expires_at,version)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
			r.ID, r.InstanceID, r.RiderID, string(r.ServiceType), string(r.Status), r.Pickup.Lat, r.Pickup.Lng, dlat, dlng, attrs, constraints, r.RequestedAt, r.ExpiresAt, r.Version)
		if err != nil {
			return err
		}
		if err := saveCommandIdempotency(ctx, tx, actorID, key, "create_request", requestHash, "request", r.ID, 201, r); err != nil {
			return err
		}
		result = r
		replay = false
		return nil
	})
	return result, replay, err
}

func (s *Store) CreateTariffIdempotent(ctx context.Context, actorID, key, requestHash string, t marketplace.DriverTariff) (marketplace.DriverTariff, bool, error) {
	if err := t.Validate(); err != nil {
		return marketplace.DriverTariff{}, false, err
	}
	var result marketplace.DriverTariff
	var replay bool
	err := s.withinSerializable(ctx, func(tx pgx.Tx) error {
		if err := lockCommandIdempotency(ctx, tx, actorID, key, "create_tariff"); err != nil {
			return err
		}
		rec, found, err := findCommandIdempotency(ctx, tx, actorID, key, "create_tariff")
		if err != nil {
			return err
		}
		if found {
			if err := decodeReplay(rec, requestHash, "tariff", &result); err != nil {
				return err
			}
			replay = true
			return nil
		}
		rules, err := json.Marshal(t.Rules)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO driver_tariffs
			(id,instance_id,driver_id,service_type,quote_mode,currency,base_fare_minor,minimum_fare_minor,per_km_minor,per_minute_minor,pickup_fee_minor,auto_quote_min_minor,auto_quote_max_minor,rules,version,active)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,TRUE)`,
			t.ID, t.InstanceID, t.DriverID, string(t.ServiceType), string(t.QuoteMode), t.BaseFare.Currency, t.BaseFare.Minor, t.MinimumFare.Minor, t.PerKM.Minor, t.PerMinute.Minor, t.PickupFee.Minor, t.AutoQuoteMinimum.Minor, t.AutoQuoteMaximum.Minor, rules, t.Version)
		if err != nil {
			return err
		}
		if err := saveCommandIdempotency(ctx, tx, actorID, key, "create_tariff", requestHash, "tariff", t.ID, 201, t); err != nil {
			return err
		}
		result = t
		replay = false
		return nil
	})
	return result, replay, err
}

func (s *Store) SubmitQuoteIdempotent(ctx context.Context, actorID, key, requestHash string, q marketplace.Quote) (marketplace.Quote, bool, error) {
	if err := q.Validate(); err != nil {
		return marketplace.Quote{}, false, err
	}
	var result marketplace.Quote
	var replay bool
	err := s.withinSerializable(ctx, func(tx pgx.Tx) error {
		if err := lockCommandIdempotency(ctx, tx, actorID, key, "submit_quote"); err != nil {
			return err
		}
		rec, found, err := findCommandIdempotency(ctx, tx, actorID, key, "submit_quote")
		if err != nil {
			return err
		}
		if found {
			if err := decodeReplay(rec, requestHash, "quote", &result); err != nil {
				return err
			}
			replay = true
			return nil
		}
		explanation, err := json.Marshal(q.Explanation)
		if err != nil {
			return err
		}
		metadata, err := json.Marshal(q.Metadata)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO marketplace_quotes
			(id,request_id,driver_id,driver_vehicle_id,tariff_id,tariff_version,status,currency,fare_minor,pickup_distance_m,pickup_eta_s,explanation,metadata,created_at,expires_at,accepted_at)
			VALUES($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,0),$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
			q.ID, q.RequestID, q.DriverID, q.DriverVehicleID, q.TariffID, q.TariffVersion, string(q.Status), q.Fare.Currency, q.Fare.Minor, q.PickupDistanceM, q.PickupETAS, explanation, metadata, q.CreatedAt, q.ExpiresAt, q.AcceptedAt)
		if err != nil {
			return err
		}
		if err := saveCommandIdempotency(ctx, tx, actorID, key, "submit_quote", requestHash, "quote", q.ID, 201, q); err != nil {
			return err
		}
		result = q
		replay = false
		return nil
	})
	return result, replay, err
}

func (s *Store) CancelRequestIdempotent(ctx context.Context, actorID, key, requestHash, requestID string) (bool, error) {
	var replay bool
	err := s.withinSerializable(ctx, func(tx pgx.Tx) error {
		if err := lockCommandIdempotency(ctx, tx, actorID, key, "cancel_request"); err != nil {
			return err
		}
		rec, found, err := findCommandIdempotency(ctx, tx, actorID, key, "cancel_request")
		if err != nil {
			return err
		}
		if found {
			if err := verifyReplay(rec, requestHash, "request"); err != nil || rec.ResourceID != requestID {
				if err != nil {
					return err
				}
				return ErrIdempotencyConflict
			}
			replay = true
			return nil
		}
		var status string
		if err := tx.QueryRow(ctx, `SELECT status FROM mobility_requests WHERE id=$1 AND rider_id=$2 FOR UPDATE`, requestID, actorID).Scan(&status); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
		if status != string(marketplace.RequestOpen) && status != string(marketplace.RequestReceivingQuotes) {
			return ErrNotFound
		}
		if _, err := tx.Exec(ctx, `UPDATE mobility_requests SET status='cancelled',version=version+1,updated_at=now() WHERE id=$1`, requestID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE marketplace_quotes SET status='invalidated' WHERE request_id=$1 AND status='pending'`, requestID); err != nil {
			return err
		}
		if err := saveCommandIdempotency(ctx, tx, actorID, key, "cancel_request", requestHash, "request", requestID, 200, map[string]any{"status": "cancelled"}); err != nil {
			return err
		}
		replay = false
		return nil
	})
	return replay, err
}

func (s *Store) WithdrawQuoteIdempotent(ctx context.Context, actorID, key, requestHash, quoteID string) (bool, error) {
	var replay bool
	err := s.withinSerializable(ctx, func(tx pgx.Tx) error {
		if err := lockCommandIdempotency(ctx, tx, actorID, key, "withdraw_quote"); err != nil {
			return err
		}
		rec, found, err := findCommandIdempotency(ctx, tx, actorID, key, "withdraw_quote")
		if err != nil {
			return err
		}
		if found {
			if err := verifyReplay(rec, requestHash, "quote"); err != nil || rec.ResourceID != quoteID {
				if err != nil {
					return err
				}
				return ErrIdempotencyConflict
			}
			replay = true
			return nil
		}
		tag, err := tx.Exec(ctx, `UPDATE marketplace_quotes SET status='withdrawn' WHERE id=$1 AND driver_id=$2 AND status='pending'`, quoteID, actorID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return ErrNotFound
		}
		if err := saveCommandIdempotency(ctx, tx, actorID, key, "withdraw_quote", requestHash, "quote", quoteID, 200, map[string]any{"status": "withdrawn"}); err != nil {
			return err
		}
		replay = false
		return nil
	})
	return replay, err
}
