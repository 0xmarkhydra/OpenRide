package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0xmarkhydra/OpenRide/packages/core-go/geo"
	"github.com/0xmarkhydra/OpenRide/packages/core-go/marketplace"
	"github.com/0xmarkhydra/OpenRide/services/marketplace/internal/app"
)

var ErrNotFound = errors.New("marketplace store: not found")

type Store struct{ pool *pgxpool.Pool }

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping pg: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}
func (s *Store) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

func (s *Store) CreateTariff(ctx context.Context, t marketplace.DriverTariff) error {
	if err := t.Validate(); err != nil {
		return err
	}
	rules, err := json.Marshal(t.Rules)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO driver_tariffs
	(id, instance_id, driver_id, service_type, quote_mode, currency, base_fare_minor, minimum_fare_minor, per_km_minor, per_minute_minor, pickup_fee_minor, auto_quote_min_minor, auto_quote_max_minor, rules, version, active)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,TRUE)`,
		t.ID, t.InstanceID, t.DriverID, string(t.ServiceType), string(t.QuoteMode), t.BaseFare.Currency, t.BaseFare.Minor, t.MinimumFare.Minor, t.PerKM.Minor, t.PerMinute.Minor, t.PickupFee.Minor, t.AutoQuoteMinimum.Minor, t.AutoQuoteMaximum.Minor, rules, t.Version)
	return err
}

func (s *Store) CreateRequest(ctx context.Context, r marketplace.Request) error {
	if err := r.Validate(); err != nil {
		return err
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
	_, err = s.pool.Exec(ctx, `INSERT INTO mobility_requests
	(id, instance_id, rider_id, service_type, status, pickup_lat, pickup_lng, destination_lat, destination_lng, attributes, constraints, requested_at, expires_at, version)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		r.ID, r.InstanceID, r.RiderID, string(r.ServiceType), string(r.Status), r.Pickup.Lat, r.Pickup.Lng, dlat, dlng, attrs, constraints, r.RequestedAt, r.ExpiresAt, r.Version)
	return err
}

func (s *Store) SubmitQuote(ctx context.Context, q marketplace.Quote) error {
	if err := q.Validate(); err != nil {
		return err
	}
	explanation, err := json.Marshal(q.Explanation)
	if err != nil {
		return err
	}
	metadata, err := json.Marshal(q.Metadata)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `INSERT INTO marketplace_quotes
	(id, request_id, driver_id, driver_vehicle_id, tariff_id, tariff_version, status, currency, fare_minor, pickup_distance_m, pickup_eta_s, explanation, metadata, created_at, expires_at, accepted_at)
	VALUES ($1,$2,$3,$4,NULLIF($5,''),NULLIF($6,0),$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		q.ID, q.RequestID, q.DriverID, q.DriverVehicleID, q.TariffID, q.TariffVersion, string(q.Status), q.Fare.Currency, q.Fare.Minor, q.PickupDistanceM, q.PickupETAS, explanation, metadata, q.CreatedAt, q.ExpiresAt, q.AcceptedAt)
	return err
}

func (s *Store) ListOffers(ctx context.Context, requestID string) ([]marketplace.Quote, error) {
	rows, err := s.pool.Query(ctx, `SELECT id,request_id,driver_id,COALESCE(driver_vehicle_id,''),COALESCE(tariff_id,''),COALESCE(tariff_version,0),status,currency,fare_minor,pickup_distance_m,pickup_eta_s,explanation,metadata,created_at,expires_at,accepted_at
	FROM marketplace_quotes WHERE request_id=$1 AND status='pending' AND expires_at>now() ORDER BY fare_minor ASC,pickup_eta_s ASC,id ASC`, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []marketplace.Quote
	for rows.Next() {
		q, err := scanQuote(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

func (s *Store) GetRequest(ctx context.Context, id string) (marketplace.Request, error) {
	return scanRequest(s.pool.QueryRow(ctx, requestSelect+` WHERE id=$1`, id))
}

func (s *Store) ListEligibleRequests(ctx context.Context, instanceID string, serviceType marketplace.ServiceType) ([]marketplace.Request, error) {
	rows, err := s.pool.Query(ctx, requestSelect+` WHERE instance_id=$1 AND service_type=$2 AND status IN ('open','receiving_quotes') AND (expires_at IS NULL OR expires_at>now()) ORDER BY requested_at ASC`, instanceID, string(serviceType))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []marketplace.Request
	for rows.Next() {
		r, err := scanRequest(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

const requestSelect = `SELECT id,instance_id,rider_id,service_type,status,pickup_lat,pickup_lng,destination_lat,destination_lng,attributes,constraints,requested_at,expires_at,version FROM mobility_requests`

type rowScanner interface{ Scan(...any) error }

func scanRequest(row rowScanner) (marketplace.Request, error) {
	var r marketplace.Request
	var service, status string
	var dlat, dlng *float64
	var attrs, constraints []byte
	err := row.Scan(&r.ID, &r.InstanceID, &r.RiderID, &service, &status, &r.Pickup.Lat, &r.Pickup.Lng, &dlat, &dlng, &attrs, &constraints, &r.RequestedAt, &r.ExpiresAt, &r.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return r, ErrNotFound
	}
	if err != nil {
		return r, err
	}
	r.ServiceType = marketplace.ServiceType(service)
	r.Status = marketplace.RequestStatus(status)
	if dlat != nil && dlng != nil {
		r.Destination = &geo.Point{Lat: *dlat, Lng: *dlng}
	}
	if len(attrs) > 0 {
		if err := json.Unmarshal(attrs, &r.Attributes); err != nil {
			return r, err
		}
	}
	if len(constraints) > 0 {
		if err := json.Unmarshal(constraints, &r.Constraints); err != nil {
			return r, err
		}
	}
	return r, nil
}

func scanQuote(row rowScanner) (marketplace.Quote, error) {
	var q marketplace.Quote
	var status, currency string
	var explanation, metadata []byte
	err := row.Scan(&q.ID, &q.RequestID, &q.DriverID, &q.DriverVehicleID, &q.TariffID, &q.TariffVersion, &status, &currency, &q.Fare.Minor, &q.PickupDistanceM, &q.PickupETAS, &explanation, &metadata, &q.CreatedAt, &q.ExpiresAt, &q.AcceptedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return q, ErrNotFound
	}
	if err != nil {
		return q, err
	}
	q.Status = marketplace.QuoteStatus(status)
	q.Fare.Currency = currency
	if len(explanation) > 0 {
		if err := json.Unmarshal(explanation, &q.Explanation); err != nil {
			return q, err
		}
	}
	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &q.Metadata); err != nil {
			return q, err
		}
	}
	return q, nil
}

func scanAgreement(row rowScanner) (marketplace.Agreement, error) {
	var a marketplace.Agreement
	var service, currency string
	var terms []byte
	err := row.Scan(&a.ID, &a.InstanceID, &a.RequestID, &a.QuoteID, &a.RiderID, &a.DriverID, &a.DriverVehicleID, &service, &currency, &a.Fare.Minor, &terms, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, ErrNotFound
	}
	if err != nil {
		return a, err
	}
	a.ServiceType = marketplace.ServiceType(service)
	a.Fare.Currency = currency
	if err := json.Unmarshal(terms, &a.TermsSnapshot); err != nil {
		return a, err
	}
	return a, nil
}

const serializableTxMaxAttempts = 3

func (s *Store) WithinTx(ctx context.Context, fn func(app.AcceptanceTx) error) error {
	var lastErr error
	for attempt := 0; attempt < serializableTxMaxAttempts; attempt++ {
		tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
		if err != nil {
			return err
		}
		err = fn(&acceptanceTx{tx: tx})
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

func isRetryableTxError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "40001" || pgErr.Code == "40P01"
}

type acceptanceTx struct{ tx pgx.Tx }

func (t *acceptanceTx) FindIdempotentAgreement(ctx context.Context, riderID, key string) (marketplace.Agreement, bool, error) {
	a, err := scanAgreement(t.tx.QueryRow(ctx, `SELECT a.id,a.instance_id,a.request_id,a.quote_id,a.rider_id,a.driver_id,COALESCE(a.driver_vehicle_id,''),a.service_type,a.currency,a.fare_minor,a.terms_snapshot,a.created_at FROM marketplace_idempotency i JOIN marketplace_agreements a ON a.id=i.agreement_id WHERE i.actor_id=$1 AND i.idempotency_key=$2 AND i.command_name='accept_quote'`, riderID, key))
	if errors.Is(err, ErrNotFound) {
		return marketplace.Agreement{}, false, nil
	}
	return a, err == nil, err
}

func (t *acceptanceTx) GetQuoteRequestID(ctx context.Context, id string) (string, error) {
	var requestID string
	err := t.tx.QueryRow(ctx, `SELECT request_id FROM marketplace_quotes WHERE id=$1`, id).Scan(&requestID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return requestID, err
}

func (t *acceptanceTx) GetRequestForUpdate(ctx context.Context, id string) (marketplace.Request, error) {
	return scanRequest(t.tx.QueryRow(ctx, requestSelect+` WHERE id=$1 FOR UPDATE`, id))
}

func (t *acceptanceTx) GetQuoteForUpdate(ctx context.Context, id string) (marketplace.Quote, error) {
	return scanQuote(t.tx.QueryRow(ctx, `SELECT id,request_id,driver_id,COALESCE(driver_vehicle_id,''),COALESCE(tariff_id,''),COALESCE(tariff_version,0),status,currency,fare_minor,pickup_distance_m,pickup_eta_s,explanation,metadata,created_at,expires_at,accepted_at FROM marketplace_quotes WHERE id=$1 FOR UPDATE`, id))
}

func (t *acceptanceTx) InsertAgreement(ctx context.Context, a marketplace.Agreement) error {
	terms, err := json.Marshal(a.TermsSnapshot)
	if err != nil {
		return err
	}
	_, err = t.tx.Exec(ctx, `INSERT INTO marketplace_agreements(id,instance_id,request_id,quote_id,rider_id,driver_id,driver_vehicle_id,service_type,currency,fare_minor,terms_snapshot,created_at) VALUES($1,$2,$3,$4,$5,$6,NULLIF($7,''),$8,$9,$10,$11,$12)`, a.ID, a.InstanceID, a.RequestID, a.QuoteID, a.RiderID, a.DriverID, a.DriverVehicleID, string(a.ServiceType), a.Fare.Currency, a.Fare.Minor, terms, a.CreatedAt)
	return err
}

func (t *acceptanceTx) MarkQuoteAccepted(ctx context.Context, id string, at time.Time) error {
	tag, err := t.tx.Exec(ctx, `UPDATE marketplace_quotes SET status='accepted',accepted_at=$2 WHERE id=$1 AND status='pending'`, id, at)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return app.ErrQuoteUnavailable
	}
	return nil
}

func (t *acceptanceTx) MarkRequestAgreed(ctx context.Context, id string) error {
	tag, err := t.tx.Exec(ctx, `UPDATE mobility_requests SET status='agreed',version=version+1,updated_at=now() WHERE id=$1 AND status IN ('open','receiving_quotes')`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return app.ErrRequestUnavailable
	}
	return nil
}

func (t *acceptanceTx) InvalidateOtherQuotes(ctx context.Context, requestID, acceptedID string) error {
	_, err := t.tx.Exec(ctx, `UPDATE marketplace_quotes SET status='invalidated' WHERE request_id=$1 AND id<>$2 AND status='pending'`, requestID, acceptedID)
	return err
}

func (t *acceptanceTx) SaveIdempotentAgreement(ctx context.Context, riderID, key, agreementID string) error {
	_, err := t.tx.Exec(ctx, `INSERT INTO marketplace_idempotency(actor_id,idempotency_key,command_name,agreement_id) VALUES($1,$2,'accept_quote',$3)`, riderID, key, agreementID)
	return err
}

func (t *acceptanceTx) AppendOutbox(ctx context.Context, e app.OutboxEvent) error {
	payload, err := json.Marshal(e.Payload)
	if err != nil {
		return err
	}
	_, err = t.tx.Exec(ctx, `INSERT INTO outbox_events(id,event_name,aggregate_type,aggregate_id,instance_id,payload,occurred_at) VALUES($1,$2,$3,$4,$5,$6,$7)`, e.ID, e.Name, e.AggregateType, e.AggregateID, e.InstanceID, payload, e.OccurredAt)
	return err
}

var _ app.AcceptanceStore = (*Store)(nil)
