package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) UpsertDevice(device Device) (Device, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	row := s.pool.QueryRow(ctx, `
		INSERT INTO push_devices (id, actor_id, actor_role, platform, token, enabled, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,TRUE,$6,$7)
		ON CONFLICT (token) DO UPDATE SET
			actor_id=EXCLUDED.actor_id,
			actor_role=EXCLUDED.actor_role,
			platform=EXCLUDED.platform,
			enabled=TRUE,
			updated_at=EXCLUDED.updated_at
		RETURNING id, actor_id, actor_role, platform, token, enabled, created_at, updated_at
	`, device.ID, device.ActorID, string(device.ActorRole), string(device.Platform), device.Token, device.CreatedAt, device.UpdatedAt)
	return scanDevice(row)
}

func (s *PostgresStore) DisableDevice(actorID string, role Role, deviceID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command, err := s.pool.Exec(ctx, `
		UPDATE push_devices SET enabled=FALSE, updated_at=NOW()
		WHERE id=$1 AND actor_id=$2 AND actor_role=$3
	`, deviceID, actorID, string(role))
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) ListDevices(actorID string, role Role) ([]Device, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT id, actor_id, actor_role, platform, token, enabled, created_at, updated_at
		FROM push_devices WHERE actor_id=$1 AND actor_role=$2 AND enabled=TRUE
		ORDER BY updated_at DESC
	`, actorID, string(role))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Device, 0)
	for rows.Next() {
		device, scanErr := scanDevice(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, device)
	}
	return items, rows.Err()
}

func (s *PostgresStore) Enqueue(message Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	payload, err := json.Marshal(message.Data)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO notification_outbox (
			id, actor_id, actor_role, event_type, resource_id, title, body, data,
			status, attempts, last_error, created_at, updated_at
		) VALUES ($1,$2,$3,$4,NULLIF($5,''),$6,$7,$8,$9,$10,NULLIF($11,''),$12,$13)
	`, message.ID, message.ActorID, string(message.ActorRole), message.EventType, message.ResourceID,
		message.Title, message.Body, payload, string(message.Status), message.Attempts, message.LastError,
		message.CreatedAt, message.UpdatedAt)
	return err
}

func (s *PostgresStore) ListPending(limit int) ([]Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, `
		SELECT id, actor_id, actor_role, event_type, COALESCE(resource_id,''), title, body,
			data, status, attempts, COALESCE(last_error,''), created_at, updated_at
		FROM notification_outbox
		WHERE status='pending'
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Message, 0)
	for rows.Next() {
		message, scanErr := scanMessage(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, message)
	}
	return items, rows.Err()
}

func (s *PostgresStore) SaveMessage(message Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command, err := s.pool.Exec(ctx, `
		UPDATE notification_outbox SET status=$2, attempts=$3, last_error=NULLIF($4,''), updated_at=$5
		WHERE id=$1
	`, message.ID, string(message.Status), message.Attempts, message.LastError, message.UpdatedAt)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface{ Scan(dest ...any) error }

func scanDevice(row scanner) (Device, error) {
	var device Device
	var role, platform string
	if err := row.Scan(&device.ID, &device.ActorID, &role, &platform, &device.Token, &device.Enabled, &device.CreatedAt, &device.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Device{}, ErrNotFound
		}
		return Device{}, err
	}
	device.ActorRole = Role(role)
	device.Platform = Platform(platform)
	return device, nil
}

func scanMessage(row scanner) (Message, error) {
	var message Message
	var role, status string
	var payload []byte
	if err := row.Scan(&message.ID, &message.ActorID, &role, &message.EventType, &message.ResourceID,
		&message.Title, &message.Body, &payload, &status, &message.Attempts, &message.LastError,
		&message.CreatedAt, &message.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Message{}, ErrNotFound
		}
		return Message{}, err
	}
	message.ActorRole = Role(role)
	message.Status = Status(status)
	if len(payload) > 0 {
		_ = json.Unmarshal(payload, &message.Data)
	}
	if message.Data == nil {
		message.Data = map[string]any{}
	}
	return message, nil
}
