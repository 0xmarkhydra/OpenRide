package custodyevidence

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) UpsertEvidence(item Evidence) (Evidence, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	row := s.pool.QueryRow(ctx, `
		INSERT INTO custody_evidence (
			id, trip_id, stage, condition_note, odometer_km, fuel_percent, battery_percent,
			driver_confirmed_at, rider_confirmed_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (trip_id, stage) DO UPDATE SET
			condition_note = EXCLUDED.condition_note,
			odometer_km = EXCLUDED.odometer_km,
			fuel_percent = EXCLUDED.fuel_percent,
			battery_percent = EXCLUDED.battery_percent,
			driver_confirmed_at = EXCLUDED.driver_confirmed_at,
			rider_confirmed_at = EXCLUDED.rider_confirmed_at,
			updated_at = EXCLUDED.updated_at
		RETURNING id, trip_id, stage, condition_note, odometer_km, fuel_percent, battery_percent,
		          driver_confirmed_at, rider_confirmed_at, created_at, updated_at
	`, item.ID, item.TripID, string(item.Stage), item.ConditionNote, item.OdometerKm,
		item.FuelPercent, item.BatteryPercent, item.DriverConfirmedAt, item.RiderConfirmedAt,
		item.CreatedAt, item.UpdatedAt)
	return scanEvidence(row)
}

func (s *PostgresStore) GetByTripStage(tripID string, stage Stage) (Evidence, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanEvidence(s.pool.QueryRow(ctx, evidenceSelect+` WHERE trip_id = $1 AND stage = $2`, tripID, string(stage)))
}

func (s *PostgresStore) ListByTrip(tripID string) ([]Evidence, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, evidenceSelect+` WHERE trip_id = $1 ORDER BY created_at ASC`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Evidence, 0, 2)
	for rows.Next() {
		item, err := scanEvidence(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *PostgresStore) SaveEvidence(item Evidence) (Evidence, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	row := s.pool.QueryRow(ctx, `
		UPDATE custody_evidence SET
			condition_note=$2, odometer_km=$3, fuel_percent=$4, battery_percent=$5,
			driver_confirmed_at=$6, rider_confirmed_at=$7, updated_at=$8
		WHERE id=$1
		RETURNING id, trip_id, stage, condition_note, odometer_km, fuel_percent, battery_percent,
		          driver_confirmed_at, rider_confirmed_at, created_at, updated_at
	`, item.ID, item.ConditionNote, item.OdometerKm, item.FuelPercent, item.BatteryPercent,
		item.DriverConfirmedAt, item.RiderConfirmedAt, item.UpdatedAt)
	return scanEvidence(row)
}

func (s *PostgresStore) CreatePhoto(photo Photo) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO custody_evidence_photos (
			id, evidence_id, photo_type, object_key, filename, content_type, size_bytes, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, photo.ID, photo.EvidenceID, photo.PhotoType, photo.ObjectKey, photo.Filename,
		photo.ContentType, photo.SizeBytes, photo.CreatedAt)
	return err
}

func (s *PostgresStore) ListPhotos(evidenceID string) ([]Photo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT id, evidence_id, photo_type, object_key, filename, content_type, size_bytes, created_at
		FROM custody_evidence_photos WHERE evidence_id = $1 ORDER BY created_at ASC
	`, evidenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Photo, 0)
	for rows.Next() {
		var photo Photo
		if err := rows.Scan(&photo.ID, &photo.EvidenceID, &photo.PhotoType, &photo.ObjectKey,
			&photo.Filename, &photo.ContentType, &photo.SizeBytes, &photo.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, photo)
	}
	return items, rows.Err()
}

const evidenceSelect = `
	SELECT id, trip_id, stage, condition_note, odometer_km, fuel_percent, battery_percent,
	       driver_confirmed_at, rider_confirmed_at, created_at, updated_at
	FROM custody_evidence`

func scanEvidence(row interface{ Scan(dest ...any) error }) (Evidence, error) {
	var item Evidence
	var stage string
	if err := row.Scan(&item.ID, &item.TripID, &stage, &item.ConditionNote, &item.OdometerKm,
		&item.FuelPercent, &item.BatteryPercent, &item.DriverConfirmedAt, &item.RiderConfirmedAt,
		&item.CreatedAt, &item.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Evidence{}, ErrNotFound
		}
		return Evidence{}, err
	}
	item.Stage = Stage(stage)
	return item, nil
}
