package payments

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(payment Payment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command, err := s.pool.Exec(ctx, `
		INSERT INTO payments (
			id, trip_id, provider, method, status, amount_minor, currency,
			provider_reference, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,NULLIF($8,''),$9,$10)
		ON CONFLICT (trip_id) DO NOTHING
	`, payment.ID, payment.TripID, payment.Provider, payment.Method, string(payment.Status),
		payment.AmountMinor, payment.Currency, payment.ProviderReference, payment.CreatedAt, payment.UpdatedAt)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrAlreadyExists
	}
	return nil
}

func (s *PostgresStore) GetByTrip(tripID string) (Payment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanPayment(s.pool.QueryRow(ctx, paymentSelect+` WHERE trip_id=$1`, tripID))
}

func (s *PostgresStore) Save(payment Payment) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command, err := s.pool.Exec(ctx, `
		UPDATE payments SET
			provider=$2,
			method=$3,
			status=$4,
			amount_minor=$5,
			currency=$6,
			provider_reference=NULLIF($7,''),
			updated_at=$8
		WHERE trip_id=$1
	`, payment.TripID, payment.Provider, payment.Method, string(payment.Status),
		payment.AmountMinor, payment.Currency, payment.ProviderReference, payment.UpdatedAt)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

const paymentSelect = `
	SELECT id, trip_id, provider, method, status, amount_minor, currency,
	       COALESCE(provider_reference,''), created_at, updated_at
	FROM payments`

func scanPayment(row interface{ Scan(dest ...any) error }) (Payment, error) {
	var payment Payment
	var status string
	if err := row.Scan(
		&payment.ID, &payment.TripID, &payment.Provider, &payment.Method, &status,
		&payment.AmountMinor, &payment.Currency, &payment.ProviderReference,
		&payment.CreatedAt, &payment.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Payment{}, ErrNotFound
		}
		return Payment{}, err
	}
	payment.Status = Status(status)
	return payment, nil
}
