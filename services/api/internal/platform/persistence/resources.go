package persistence

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"flashx/services/api/internal/platform/cache"
	"flashx/services/api/internal/platform/config"
	"flashx/services/api/internal/platform/database"
)

type Resources struct {
	Postgres *pgxpool.Pool
	Redis    *redis.Client
}

func Open(ctx context.Context, cfg config.Config) (*Resources, error) {
	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	redisClient, err := cache.Open(ctx, cfg.RedisAddr, cfg.RedisPassword)
	if err != nil {
		pool.Close()
		return nil, err
	}
	return &Resources{Postgres: pool, Redis: redisClient}, nil
}

func (r *Resources) Close() {
	if r == nil {
		return
	}
	if r.Redis != nil {
		_ = r.Redis.Close()
	}
	if r.Postgres != nil {
		r.Postgres.Close()
	}
}

func (r *Resources) Ready(ctx context.Context) error {
	if r == nil || r.Postgres == nil || r.Redis == nil {
		return fmt.Errorf("persistence resources are not initialized")
	}
	if err := r.Postgres.Ping(ctx); err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	if err := r.Redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	return nil
}
