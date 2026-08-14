package operationalsettings

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) Get() (Config, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return scanConfig(s.pool.QueryRow(ctx, settingsSelect+` WHERE id='default'`))
}

func (s *PostgresStore) Save(config Config) (Config, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	row := s.pool.QueryRow(ctx, `
		INSERT INTO operational_settings (
			id, designated_driver_car_enabled, designated_driver_bike_enabled,
			vehicle_inspection_assist_enabled, dispatch_max_distance_m,
			driver_location_max_age_seconds, version, updated_by, updated_at
		) VALUES ('default',$1,$2,$3,$4,$5,1,NULLIF($6,''),$7)
		ON CONFLICT (id) DO UPDATE SET
			designated_driver_car_enabled=EXCLUDED.designated_driver_car_enabled,
			designated_driver_bike_enabled=EXCLUDED.designated_driver_bike_enabled,
			vehicle_inspection_assist_enabled=EXCLUDED.vehicle_inspection_assist_enabled,
			dispatch_max_distance_m=EXCLUDED.dispatch_max_distance_m,
			driver_location_max_age_seconds=EXCLUDED.driver_location_max_age_seconds,
			version=operational_settings.version+1,
			updated_by=EXCLUDED.updated_by,
			updated_at=EXCLUDED.updated_at
		RETURNING designated_driver_car_enabled, designated_driver_bike_enabled,
			vehicle_inspection_assist_enabled, dispatch_max_distance_m,
			driver_location_max_age_seconds, version, COALESCE(updated_by,''), updated_at
	`, config.DesignatedDriverCarEnabled, config.DesignatedDriverBikeEnabled,
		config.VehicleInspectionEnabled, config.DispatchMaxDistanceM,
		config.DriverLocationMaxAgeSeconds, config.UpdatedBy, config.UpdatedAt)
	return scanConfigValues(row)
}

const settingsSelect = `SELECT designated_driver_car_enabled, designated_driver_bike_enabled,
	vehicle_inspection_assist_enabled, dispatch_max_distance_m,
	driver_location_max_age_seconds, version, COALESCE(updated_by,''), updated_at
	FROM operational_settings`

type configScanner interface{ Scan(dest ...any) error }

func scanConfig(row configScanner) (Config, error) {
	cfg, err := scanConfigValues(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Config{}, ErrNotFound
	}
	return cfg, err
}

func scanConfigValues(row configScanner) (Config, error) {
	var cfg Config
	err := row.Scan(&cfg.DesignatedDriverCarEnabled, &cfg.DesignatedDriverBikeEnabled,
		&cfg.VehicleInspectionEnabled, &cfg.DispatchMaxDistanceM,
		&cfg.DriverLocationMaxAgeSeconds, &cfg.Version, &cfg.UpdatedBy, &cfg.UpdatedAt)
	return cfg, err
}
