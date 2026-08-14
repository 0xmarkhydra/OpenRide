package pricing

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) GetActive(serviceType string) (RuleRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanRule(s.pool.QueryRow(ctx, pricingSelect+` WHERE service_type=$1 AND active=TRUE ORDER BY effective_at DESC, created_at DESC LIMIT 1`, serviceType))
}

func (s *PostgresStore) ListActive() ([]RuleRecord, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, pricingSelect+` WHERE active=TRUE ORDER BY service_type, effective_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]RuleRecord, 0)
	for rows.Next() {
		item, scanErr := scanRule(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PostgresStore) ReplaceActive(item RuleRecord) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, item.ServiceType); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE pricing_rules SET active=FALSE, updated_at=NOW() WHERE service_type=$1 AND active=TRUE`, item.ServiceType); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO pricing_rules (
			id, service_type, base_fare_minor, per_km_minor, per_minute_minor,
			service_minor, minimum_fare_minor, currency, active, effective_at,
			pricing_version, updated_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,0,$5,$6,$7,TRUE,$8,$9,NULLIF($10,''),NOW(),NOW())
	`, item.ID, item.ServiceType, item.BaseFareMinor, item.PerKMMinor, item.ServiceMinor,
		item.MinimumMinor, item.Currency, item.EffectiveAt, item.PricingVersion, item.UpdatedBy)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

const pricingSelect = `SELECT id, service_type, base_fare_minor, per_km_minor,
	COALESCE(service_minor,0), minimum_fare_minor, currency,
	COALESCE(NULLIF(pricing_version,''), id), effective_at, COALESCE(updated_by,'')
	FROM pricing_rules`

type ruleScanner interface{ Scan(dest ...any) error }

func scanRule(row ruleScanner) (RuleRecord, error) {
	var item RuleRecord
	if err := row.Scan(&item.ID, &item.ServiceType, &item.BaseFareMinor, &item.PerKMMinor,
		&item.ServiceMinor, &item.MinimumMinor, &item.Currency, &item.PricingVersion,
		&item.EffectiveAt, &item.UpdatedBy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RuleRecord{}, ErrRuleNotFound
		}
		return RuleRecord{}, err
	}
	return item, nil
}
