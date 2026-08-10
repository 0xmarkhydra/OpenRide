package trips

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

func (s *PostgresStore) Create(trip Trip) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO trips (
			id, rider_id, driver_id, service_type, status,
			pickup, destination, estimated_distance_m, estimated_duration_s,
			estimated_fare_minor, final_fare_minor, base_fare_minor,
			distance_fare_minor, discount_minor, currency, created_at,
			accepted_at, arrived_at, started_at, completed_at, cancelled_at,
			cancellation_reason, version
		) VALUES (
			$1, $2, NULLIF($3, ''), $4, $5,
			ST_SetSRID(ST_MakePoint($6, $7), 4326)::geoflashxhy,
			ST_SetSRID(ST_MakePoint($8, $9), 4326)::geoflashxhy,
			$10, $11, $12, $13, $14, $15, $16, $17, $18,
			$19, $20, $21, $22, $23, $24, $25
		)
	`, trip.ID, trip.RiderID, trip.DriverID, trip.ServiceType, string(trip.Status),
		trip.Pickup.Lng, trip.Pickup.Lat, trip.Destination.Lng, trip.Destination.Lat,
		trip.EstimatedDistanceM, trip.EstimatedDurationS, trip.EstimatedFareMinor,
		trip.FinalFareMinor, trip.FareBreakdown.BaseFareMinor, trip.FareBreakdown.DistanceMinor,
		trip.FareBreakdown.DiscountMinor, trip.Currency, trip.CreatedAt, trip.AcceptedAt,
		trip.ArrivedAt, trip.StartedAt, trip.CompletedAt, trip.CancelledAt,
		trip.CancellationReason, trip.Version)
	return err
}

func (s *PostgresStore) Get(id string) (Trip, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanTrip(s.pool.QueryRow(ctx, tripSelect+` WHERE id = $1`, id))
}

func (s *PostgresStore) Save(trip Trip) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	command, err := s.pool.Exec(ctx, `
		UPDATE trips SET
			driver_id = NULLIF($2, ''),
			service_type = $3,
			status = $4,
			estimated_distance_m = $5,
			estimated_duration_s = $6,
			estimated_fare_minor = $7,
			final_fare_minor = $8,
			base_fare_minor = $9,
			distance_fare_minor = $10,
			discount_minor = $11,
			currency = $12,
			accepted_at = $13,
			arrived_at = $14,
			started_at = $15,
			completed_at = $16,
			cancelled_at = $17,
			cancellation_reason = $18,
			version = version + 1
		WHERE id = $1 AND version = $19
	`, trip.ID, trip.DriverID, trip.ServiceType, string(trip.Status),
		trip.EstimatedDistanceM, trip.EstimatedDurationS, trip.EstimatedFareMinor,
		trip.FinalFareMinor, trip.FareBreakdown.BaseFareMinor, trip.FareBreakdown.DistanceMinor,
		trip.FareBreakdown.DiscountMinor, trip.Currency, trip.AcceptedAt, trip.ArrivedAt,
		trip.StartedAt, trip.CompletedAt, trip.CancelledAt, trip.CancellationReason, trip.Version)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 1 {
		return nil
	}

	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM trips WHERE id = $1)`, trip.ID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return ErrVersionConflict
}

func (s *PostgresStore) ListByRider(riderID string, limit int) ([]Trip, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, tripSelect+` WHERE rider_id = $1 ORDER BY created_at DESC LIMIT $2`, riderID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectTrips(rows)
}

func (s *PostgresStore) ListByDriver(driverID string, limit int) ([]Trip, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, tripSelect+` WHERE driver_id = $1 ORDER BY created_at DESC LIMIT $2`, driverID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectTrips(rows)
}

func (s *PostgresStore) ListSearching(limit int) ([]Trip, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, tripSelect+` WHERE status = 'searching' AND driver_id IS NULL ORDER BY created_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectTrips(rows)
}

func (s *PostgresStore) ListAll(limit int) ([]Trip, error) {
	if limit <= 0 || limit > 500 { limit = 100 }
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, tripSelect+` ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil { return nil, err }
	defer rows.Close()
	return collectTrips(rows)
}

func (s *PostgresStore) FindActiveByDriver(driverID string) (Trip, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanTrip(s.pool.QueryRow(ctx, tripSelect+` WHERE driver_id=$1 AND status IN ('accepted','arriving','arrived','in_progress') ORDER BY created_at DESC LIMIT 1`, driverID))
}

const tripSelect = `
	SELECT
		id::text,
		rider_id::text,
		COALESCE(driver_id::text, ''),
		service_type,
		status,
		ST_Y(pickup::geometry), ST_X(pickup::geometry),
		ST_Y(destination::geometry), ST_X(destination::geometry),
		COALESCE(estimated_distance_m, 0),
		COALESCE(estimated_duration_s, 0),
		COALESCE(estimated_fare_minor, 0),
		COALESCE(final_fare_minor, 0),
		COALESCE(base_fare_minor, 0),
		COALESCE(distance_fare_minor, 0),
		COALESCE(discount_minor, 0),
		currency,
		created_at,
		accepted_at,
		arrived_at,
		started_at,
		completed_at,
		cancelled_at,
		COALESCE(cancellation_reason, ''),
		version
	FROM trips`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTrip(row rowScanner) (Trip, error) {
	var trip Trip
	var status string
	var pickupLat, pickupLng, destinationLat, destinationLng float64
	if err := row.Scan(
		&trip.ID, &trip.RiderID, &trip.DriverID, &trip.ServiceType, &status,
		&pickupLat, &pickupLng, &destinationLat, &destinationLng,
		&trip.EstimatedDistanceM, &trip.EstimatedDurationS, &trip.EstimatedFareMinor,
		&trip.FinalFareMinor, &trip.FareBreakdown.BaseFareMinor, &trip.FareBreakdown.DistanceMinor,
		&trip.FareBreakdown.DiscountMinor, &trip.Currency, &trip.CreatedAt, &trip.AcceptedAt,
		&trip.ArrivedAt, &trip.StartedAt, &trip.CompletedAt, &trip.CancelledAt,
		&trip.CancellationReason, &trip.Version,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Trip{}, ErrNotFound
		}
		return Trip{}, err
	}
	trip.Status = Status(status)
	trip.Pickup = Point{Lat: pickupLat, Lng: pickupLng}
	trip.Destination = Point{Lat: destinationLat, Lng: destinationLng}
	trip.FareBreakdown.TotalMinor = trip.EstimatedFareMinor
	return trip, nil
}

func collectTrips(rows pgx.Rows) ([]Trip, error) {
	result := make([]Trip, 0)
	for rows.Next() {
		trip, err := scanTrip(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, trip)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
