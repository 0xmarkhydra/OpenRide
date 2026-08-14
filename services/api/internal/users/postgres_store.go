package users

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(user User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO users (id, phone, full_name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.Phone, user.FullName, user.Status, user.CreatedAt, user.UpdatedAt)
	if isUniqueViolation(err) {
		return ErrAlreadyExists
	}
	return err
}

func (s *PostgresStore) Get(id string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanUser(s.pool.QueryRow(ctx, userSelect+` WHERE id = $1`, id))
}

func (s *PostgresStore) GetByPhone(phone string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanUser(s.pool.QueryRow(ctx, userSelect+` WHERE phone = $1`, phone))
}

func (s *PostgresStore) ListAll(limit int) ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, userSelect+` ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]User, 0)
	for rows.Next() {
		user, scanErr := scanUser(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, user)
	}
	return items, rows.Err()
}

func (s *PostgresStore) Save(user User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command, err := s.pool.Exec(ctx, `
		UPDATE users SET phone=$2, full_name=$3, status=$4, updated_at=$5 WHERE id=$1
	`, user.ID, user.Phone, user.FullName, user.Status, user.UpdatedAt)
	if isUniqueViolation(err) {
		return ErrAlreadyExists
	}
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

const userSelect = `SELECT id, phone, full_name, status, created_at, updated_at FROM users`

func scanUser(row interface{ Scan(dest ...any) error }) (User, error) {
	var user User
	if err := row.Scan(&user.ID, &user.Phone, &user.FullName, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
