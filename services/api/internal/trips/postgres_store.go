package trips

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) Create(trip Trip) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO trips (
			id, rider_id, driver_id, customer_vehicle_id, service_type, booking_mode, scheduled_at, status,
			pickup_lat, pickup_lng, destination_lat, destination_lng, estimated_distance_m, estimated_duration_s,
			estimated_fare_minor, final_fare_minor, base_fare_minor, distance_fare_minor,
			service_fare_minor, schedule_fare_minor, surcharge_minor, discount_minor, currency, created_at,
			accepted_at, arrived_at, vehicle_received_at, started_at, handover_at, completed_at, cancelled_at,
			cancellation_reason, inspection_result, incident_type, incident_note, incident_open, version
		) VALUES (
			$1,$2,NULLIF($3,''),NULLIF($4,''),$5,$6,$7,$8,
			$9,$10,$11,$12,
			$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35,$36,$37
		)
	`, trip.ID, trip.RiderID, trip.DriverID, trip.CustomerVehicleID, trip.ServiceType, trip.BookingMode, trip.ScheduledAt, string(trip.Status),
		trip.Pickup.Lat, trip.Pickup.Lng, trip.Destination.Lat, trip.Destination.Lng,
		trip.EstimatedDistanceM, trip.EstimatedDurationS, trip.EstimatedFareMinor, trip.FinalFareMinor,
		trip.FareBreakdown.BaseFareMinor, trip.FareBreakdown.DistanceMinor, trip.FareBreakdown.ServiceMinor,
		trip.FareBreakdown.ScheduleMinor, trip.FareBreakdown.SurchargeMinor, trip.FareBreakdown.DiscountMinor,
		trip.Currency, trip.CreatedAt, trip.AcceptedAt, trip.ArrivedAt, trip.VehicleReceivedAt, trip.StartedAt,
		trip.HandoverAt, trip.CompletedAt, trip.CancelledAt, trip.CancellationReason, trip.InspectionResult,
		trip.IncidentType, trip.IncidentNote, trip.IncidentOpen, trip.Version)
	return err
}

func (s *PostgresStore) Get(id string) (Trip, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanTrip(s.pool.QueryRow(ctx, tripSelect+` WHERE id=$1`, id))
}

func (s *PostgresStore) Save(trip Trip) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command, err := s.pool.Exec(ctx, `
		UPDATE trips SET
			driver_id=NULLIF($2,''), customer_vehicle_id=NULLIF($3,''), service_type=$4,
			booking_mode=$5, scheduled_at=$6, status=$7,
			estimated_distance_m=$8, estimated_duration_s=$9, estimated_fare_minor=$10, final_fare_minor=$11,
			base_fare_minor=$12, distance_fare_minor=$13, service_fare_minor=$14, schedule_fare_minor=$15,
			surcharge_minor=$16, discount_minor=$17, currency=$18,
			accepted_at=$19, arrived_at=$20, vehicle_received_at=$21, started_at=$22, handover_at=$23,
			completed_at=$24, cancelled_at=$25, cancellation_reason=$26, inspection_result=$27,
			incident_type=$28, incident_note=$29, incident_open=$30, version=version+1
		WHERE id=$1 AND version=$31
	`, trip.ID, trip.DriverID, trip.CustomerVehicleID, trip.ServiceType, trip.BookingMode, trip.ScheduledAt, string(trip.Status),
		trip.EstimatedDistanceM, trip.EstimatedDurationS, trip.EstimatedFareMinor, trip.FinalFareMinor,
		trip.FareBreakdown.BaseFareMinor, trip.FareBreakdown.DistanceMinor, trip.FareBreakdown.ServiceMinor,
		trip.FareBreakdown.ScheduleMinor, trip.FareBreakdown.SurchargeMinor, trip.FareBreakdown.DiscountMinor,
		trip.Currency, trip.AcceptedAt, trip.ArrivedAt, trip.VehicleReceivedAt, trip.StartedAt, trip.HandoverAt,
		trip.CompletedAt, trip.CancelledAt, trip.CancellationReason, trip.InspectionResult, trip.IncidentType,
		trip.IncidentNote, trip.IncidentOpen, trip.Version)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 1 {
		return nil
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM trips WHERE id=$1)`, trip.ID).Scan(&exists); err != nil {
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
	rows, err := s.pool.Query(ctx, tripSelect+` WHERE rider_id=$1 ORDER BY created_at DESC LIMIT $2`, riderID, limit)
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
	rows, err := s.pool.Query(ctx, tripSelect+` WHERE driver_id=$1 ORDER BY created_at DESC LIMIT $2`, driverID, limit)
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
	rows, err := s.pool.Query(ctx, tripSelect+` WHERE status='searching' AND driver_id IS NULL ORDER BY created_at ASC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectTrips(rows)
}
func (s *PostgresStore) ListScheduledDue(before time.Time, limit int) ([]Trip, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, tripSelect+` WHERE status='scheduled' AND scheduled_at IS NOT NULL AND scheduled_at <= $1 ORDER BY scheduled_at ASC LIMIT $2`, before, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectTrips(rows)
}
func (s *PostgresStore) ListAll(limit int) ([]Trip, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, tripSelect+` ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectTrips(rows)
}
func (s *PostgresStore) FindActiveByDriver(driverID string) (Trip, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanTrip(s.pool.QueryRow(ctx, tripSelect+` WHERE driver_id=$1 AND status NOT IN ('scheduled','searching','completed','cancelled') ORDER BY created_at DESC LIMIT 1`, driverID))
}

const tripSelect = `SELECT
	id::text, rider_id::text, COALESCE(driver_id::text,''), COALESCE(customer_vehicle_id::text,''),
	service_type, booking_mode, scheduled_at, status,
	pickup_lat, pickup_lng, destination_lat, destination_lng,
	COALESCE(estimated_distance_m,0), COALESCE(estimated_duration_s,0), COALESCE(estimated_fare_minor,0),
	COALESCE(final_fare_minor,0), COALESCE(base_fare_minor,0), COALESCE(distance_fare_minor,0),
	COALESCE(service_fare_minor,0), COALESCE(schedule_fare_minor,0), COALESCE(surcharge_minor,0), COALESCE(discount_minor,0),
	currency, created_at, accepted_at, arrived_at, vehicle_received_at, started_at, handover_at, completed_at, cancelled_at,
	COALESCE(cancellation_reason,''), COALESCE(inspection_result,''), COALESCE(incident_type,''), COALESCE(incident_note,''), COALESCE(incident_open,FALSE), version
	FROM trips`

type rowScanner interface{ Scan(dest ...any) error }

func scanTrip(row rowScanner) (Trip, error) {
	var trip Trip
	var status string
	var pickupLat, pickupLng, destinationLat, destinationLng float64
	if err := row.Scan(&trip.ID, &trip.RiderID, &trip.DriverID, &trip.CustomerVehicleID,
		&trip.ServiceType, &trip.BookingMode, &trip.ScheduledAt, &status,
		&pickupLat, &pickupLng, &destinationLat, &destinationLng,
		&trip.EstimatedDistanceM, &trip.EstimatedDurationS, &trip.EstimatedFareMinor, &trip.FinalFareMinor,
		&trip.FareBreakdown.BaseFareMinor, &trip.FareBreakdown.DistanceMinor, &trip.FareBreakdown.ServiceMinor,
		&trip.FareBreakdown.ScheduleMinor, &trip.FareBreakdown.SurchargeMinor, &trip.FareBreakdown.DiscountMinor,
		&trip.Currency, &trip.CreatedAt, &trip.AcceptedAt, &trip.ArrivedAt, &trip.VehicleReceivedAt, &trip.StartedAt,
		&trip.HandoverAt, &trip.CompletedAt, &trip.CancelledAt, &trip.CancellationReason, &trip.InspectionResult,
		&trip.IncidentType, &trip.IncidentNote, &trip.IncidentOpen, &trip.Version); err != nil {
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
