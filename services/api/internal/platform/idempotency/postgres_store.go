package idempotency

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
	ttl  time.Duration
}

func NewPostgresStore(pool *pgxpool.Pool, ttl time.Duration) *PostgresStore {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &PostgresStore{pool: pool, ttl: ttl}
}

func (s *PostgresStore) Get(scope, key string) (Record, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	var record Record
	err := s.pool.QueryRow(ctx, `
		SELECT fingerprint, resource_id
		FROM idempotency_records
		WHERE scope = $1 AND key = $2 AND expires_at > NOW()
	`, scope, key).Scan(&record.Fingerprint, &record.ResourceID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, err
	}
	return record, true, nil
}

func (s *PostgresStore) Put(scope, key string, record Record) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	command, err := s.pool.Exec(ctx, `
		INSERT INTO idempotency_records (scope, key, fingerprint, resource_id, expires_at)
		VALUES ($1, $2, $3, $4, NOW() + $5::interval)
		ON CONFLICT (scope, key) DO NOTHING
	`, scope, key, record.Fingerprint, record.ResourceID, intervalLiteral(s.ttl))
	if err != nil {
		return err
	}
	if command.RowsAffected() == 1 {
		return nil
	}
	current, ok, err := s.Get(scope, key)
	if err != nil {
		return err
	}
	if !ok || current.Fingerprint != record.Fingerprint || current.ResourceID != record.ResourceID {
		return ErrConflict
	}
	return nil
}

func intervalLiteral(duration time.Duration) string {
	seconds := int64(duration.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return time.Duration(seconds * int64(time.Second)).String()
}
