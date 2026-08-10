package ratings

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) Create(rating Rating) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command, err := s.pool.Exec(ctx, `
		INSERT INTO ratings (id, trip_id, rider_id, driver_id, stars, comment, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (trip_id) DO NOTHING
	`, rating.ID, rating.TripID, rating.RiderID, rating.DriverID, rating.Stars, rating.Comment, rating.CreatedAt)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrAlreadyExists
	}
	return nil
}

func (s *PostgresStore) GetByTrip(tripID string) (Rating, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var rating Rating
	err := s.pool.QueryRow(ctx, `
		SELECT id, trip_id, rider_id, driver_id, stars, comment, created_at
		FROM ratings WHERE trip_id=$1
	`, tripID).Scan(&rating.ID, &rating.TripID, &rating.RiderID, &rating.DriverID, &rating.Stars, &rating.Comment, &rating.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Rating{}, ErrNotFound
	}
	return rating, err
}
