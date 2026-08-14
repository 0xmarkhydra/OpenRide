package admin

import (
	"context"
	"encoding/json"
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

func (s *PostgresStore) Bootstrap(user User) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO admin_users (id, phone, email, display_name, role, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (phone) WHERE phone IS NOT NULL DO NOTHING
	`, user.ID, user.Phone, user.Email, user.DisplayName, user.Role, user.Status, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return User{}, translateAdminStoreError(err)
	}
	return s.GetByPhone(user.Phone)
}

func (s *PostgresStore) Create(user User) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, err := s.pool.Exec(ctx, `
		INSERT INTO admin_users (id, phone, email, display_name, role, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, user.ID, user.Phone, user.Email, user.DisplayName, user.Role, user.Status, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return User{}, translateAdminStoreError(err)
	}
	return s.Get(user.ID)
}

func (s *PostgresStore) Save(user User) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	tag, err := s.pool.Exec(ctx, `
		UPDATE admin_users
		SET email=$2, display_name=$3, role=$4, status=$5, updated_at=$6
		WHERE id=$1
	`, user.ID, user.Email, user.DisplayName, user.Role, user.Status, user.UpdatedAt)
	if err != nil {
		return User{}, translateAdminStoreError(err)
	}
	if tag.RowsAffected() == 0 {
		return User{}, ErrNotFound
	}
	return s.Get(user.ID)
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

func (s *PostgresStore) List(limit int) ([]User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, adminSelect+` ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]User, 0)
	for rows.Next() {
		item, scanErr := scanAdmin(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
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

func (s *PostgresStore) ListAudit(limit int) ([]AuditEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.pool.Query(ctx, `
		SELECT actor_type, COALESCE(actor_id,''), action, resource_type, COALESCE(resource_id,''), metadata, created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]AuditEntry, 0)
	for rows.Next() {
		var entry AuditEntry
		var rawMetadata []byte
		if err := rows.Scan(&entry.ActorType, &entry.ActorID, &entry.Action, &entry.ResourceType, &entry.ResourceID, &rawMetadata, &entry.CreatedAt); err != nil {
			return nil, err
		}
		if len(rawMetadata) > 0 {
			_ = json.Unmarshal(rawMetadata, &entry.Metadata)
		}
		items = append(items, entry)
	}
	return items, rows.Err()
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

func translateAdminStoreError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
}
