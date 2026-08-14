package customervehicles

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) Create(vehicle Vehicle) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO customer_vehicles (
			id, owner_user_id, type, license_plate, brand, model, year, color,
			transmission, seats, notes, photo_object_key, status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,0),$8,$9,NULLIF($10,0),$11,$12,$13,$14,$15)
	`, vehicle.ID, vehicle.OwnerUserID, vehicle.Type, vehicle.LicensePlate, vehicle.Brand, vehicle.Model,
		vehicle.Year, vehicle.Color, vehicle.Transmission, vehicle.Seats, vehicle.Notes, vehicle.PhotoObjectKey,
		vehicle.Status, vehicle.CreatedAt, vehicle.UpdatedAt)
	return err
}

func (s *PostgresStore) Get(id string) (Vehicle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanVehicle(s.pool.QueryRow(ctx, vehicleSelect+` WHERE id=$1`, id))
}

func (s *PostgresStore) ListByOwner(ownerUserID string) ([]Vehicle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, vehicleSelect+` WHERE owner_user_id=$1 ORDER BY created_at DESC`, ownerUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Vehicle, 0)
	for rows.Next() {
		vehicle, scanErr := scanVehicle(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, vehicle)
	}
	return items, rows.Err()
}

func (s *PostgresStore) ListAll(limit int) ([]Vehicle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, vehicleSelect+` ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Vehicle, 0)
	for rows.Next() {
		vehicle, scanErr := scanVehicle(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, vehicle)
	}
	return items, rows.Err()
}

const vehicleSelect = `SELECT id, owner_user_id, type, license_plate, brand, model,
	COALESCE(year,0), color, transmission, COALESCE(seats,0), notes,
	photo_object_key, status, created_at, updated_at FROM customer_vehicles`

func scanVehicle(row interface{ Scan(dest ...any) error }) (Vehicle, error) {
	var vehicle Vehicle
	if err := row.Scan(&vehicle.ID, &vehicle.OwnerUserID, &vehicle.Type, &vehicle.LicensePlate,
		&vehicle.Brand, &vehicle.Model, &vehicle.Year, &vehicle.Color, &vehicle.Transmission,
		&vehicle.Seats, &vehicle.Notes, &vehicle.PhotoObjectKey, &vehicle.Status, &vehicle.CreatedAt, &vehicle.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Vehicle{}, ErrNotFound
		}
		return Vehicle{}, err
	}
	return vehicle, nil
}
