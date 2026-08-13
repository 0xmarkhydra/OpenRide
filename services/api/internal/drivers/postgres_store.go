package drivers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) Create(driver Driver) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	phone := driver.Phone
	if phone == "" {
		phone = syntheticPhone(driver.ID)
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO drivers (
			id, phone, full_name, approval_status, availability_status,
			service_type, capabilities, last_idle_at, created_at, updated_at, version
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, driver.ID, phone, driver.FullName, string(driver.Approval), string(driver.Availability),
		driver.ServiceType, driver.Capabilities, driver.LastIdleAt, driver.CreatedAt, driver.UpdatedAt, driver.Version)
	return err
}

func (s *PostgresStore) Get(id string) (Driver, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanDriver(s.pool.QueryRow(ctx, driverSelect+` WHERE id=$1`, id))
}
func (s *PostgresStore) GetByPhone(phone string) (Driver, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanDriver(s.pool.QueryRow(ctx, driverSelect+` WHERE phone=$1`, phone))
}

func (s *PostgresStore) Save(driver Driver) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	command, err := s.pool.Exec(ctx, `
		UPDATE drivers SET phone=$2, full_name=$3, approval_status=$4, availability_status=$5,
			service_type=$6, capabilities=$7, last_idle_at=$8, updated_at=$9, version=version+1
		WHERE id=$1 AND version=$10
	`, driver.ID, driver.Phone, driver.FullName, string(driver.Approval), string(driver.Availability),
		driver.ServiceType, driver.Capabilities, driver.LastIdleAt, driver.UpdatedAt, driver.Version)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 1 {
		return nil
	}
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM drivers WHERE id=$1)`, driver.ID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return ErrDriverUnavailable
}

func (s *PostgresStore) TryMarkBusy(id string) (Driver, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	row := s.pool.QueryRow(ctx, `
		UPDATE drivers SET availability_status='busy', updated_at=NOW(), version=version+1
		WHERE id=$1 AND approval_status='approved' AND availability_status='online'
		RETURNING id::text, COALESCE(phone,''), full_name, service_type, capabilities, approval_status,
			availability_status, last_idle_at, created_at, updated_at, version
	`, id)
	driver, err := scanDriver(row)
	if errors.Is(err, ErrNotFound) {
		if _, getErr := s.Get(id); getErr != nil {
			return Driver{}, getErr
		}
		return Driver{}, ErrDriverUnavailable
	}
	return driver, err
}

func (s *PostgresStore) List() ([]Driver, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := s.pool.Query(ctx, driverSelect+` ORDER BY created_at DESC LIMIT 1000`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Driver, 0)
	for rows.Next() {
		driver, scanErr := scanDriver(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, driver)
	}
	return result, rows.Err()
}

const driverSelect = `SELECT id::text, COALESCE(phone,''), full_name, service_type, capabilities, approval_status,
	availability_status, last_idle_at, created_at, updated_at, version FROM drivers`

func scanDriver(row interface{ Scan(dest ...any) error }) (Driver, error) {
	var driver Driver
	var approval, availability string
	if err := row.Scan(&driver.ID, &driver.Phone, &driver.FullName, &driver.ServiceType, &driver.Capabilities,
		&approval, &availability, &driver.LastIdleAt, &driver.CreatedAt, &driver.UpdatedAt, &driver.Version); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Driver{}, ErrNotFound
		}
		return Driver{}, err
	}
	driver.Approval = ApprovalStatus(approval)
	driver.Availability = AvailabilityStatus(availability)
	return driver, nil
}

func syntheticPhone(driverID string) string {
	sum := sha256.Sum256([]byte(driverID))
	return "dev-" + hex.EncodeToString(sum[:12])
}
