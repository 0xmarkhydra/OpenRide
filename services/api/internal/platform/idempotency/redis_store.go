package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client, ttl: 24 * time.Hour}
}

func (s *RedisStore) Get(scope, key string) (Record, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	payload, err := s.client.Get(ctx, redisKey(scope, key)).Bytes()
	if errors.Is(err, redis.Nil) { return Record{}, false, nil }
	if err != nil { return Record{}, false, err }
	var record Record
	if err := json.Unmarshal(payload, &record); err != nil { return Record{}, false, err }
	return record, true, nil
}

func (s *RedisStore) Put(scope, key string, record Record) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	payload, err := json.Marshal(record)
	if err != nil { return err }
	created, err := s.client.SetNX(ctx, redisKey(scope, key), payload, s.ttl).Result()
	if err != nil { return err }
	if created { return nil }
	current, exists, err := s.Get(scope, key)
	if err != nil { return err }
	if !exists || current.Fingerprint != record.Fingerprint { return ErrConflict }
	return nil
}

func redisKey(scope, key string) string { return "idempotency:" + scope + ":" + key }
