package admin

import (
	"context"
	"encoding/json"
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

func (s *PostgresStore) Bootstrap(user User) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO admin_users (id, phone, email, display_name, role, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (phone) WHERE phone IS NOT NULL DO NOTHING
	`, user.ID, user.Phone, user.Email, user.DisplayName, user.Role, user.Status, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return User{}, err
	}
	return s.GetByPhone(user.Phone)
}

func (s *PostgresStore) Get(id string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanAdmin(s.pool.QueryRow(ctx, adminSelect+` WHERE id=$1`, id))
}

func (s *PostgresStore) GetByPhone(phone string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanAdmin(s.pool.QueryRow(ctx, adminSelect+` WHERE phone=$1`, phone))
}

func (s *PostgresStore) AppendAudit(entry AuditEntry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	metadata, err := json.Marshal(entry.Metadata)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO audit_logs (actor_type, actor_id, action, resource_type, resource_id, metadata, created_at)
		VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7)
	`, entry.ActorType, entry.ActorID, entry.Action, entry.ResourceType, entry.ResourceID, string(metadata), entry.CreatedAt)
	return err
}

const adminSelect = `
	SELECT id, COALESCE(phone,''), email, display_name, role, status, created_at, updated_at
	FROM admin_users`

func scanAdmin(row interface{ Scan(dest ...any) error }) (User, error) {
	var user User
	if err := row.Scan(&user.ID, &user.Phone, &user.Email, &user.DisplayName, &user.Role, &user.Status, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return user, nil
}
