package dispatch

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Locker serializes mutations for one trip across API instances. A failed
// acquisition is intentionally non-blocking so callers can retry/recover from
// the durable trip state instead of holding HTTP requests indefinitely.
type Locker interface {
	TryLock(key string, ttl time.Duration) (unlock func(), acquired bool, err error)
}

type MemoryLocker struct {
	mu    sync.Mutex
	locks map[string]bool
}

func NewMemoryLocker() *MemoryLocker {
	return &MemoryLocker{locks: make(map[string]bool)}
}

func (l *MemoryLocker) TryLock(key string, _ time.Duration) (func(), bool, error) {
	l.mu.Lock()
	if l.locks[key] {
		l.mu.Unlock()
		return nil, false, nil
	}
	l.locks[key] = true
	l.mu.Unlock()
	var once sync.Once
	return func() {
		once.Do(func() {
			l.mu.Lock()
			delete(l.locks, key)
			l.mu.Unlock()
		})
	}, true, nil
}

type RedisLocker struct {
	client *redis.Client
	prefix string
}

func NewRedisLocker(client *redis.Client, prefix string) *RedisLocker {
	if prefix == "" {
		prefix = "flashx"
	}
	return &RedisLocker{client: client, prefix: prefix}
}

func (l *RedisLocker) TryLock(key string, ttl time.Duration) (func(), bool, error) {
	if ttl <= 0 {
		ttl = 3 * time.Second
	}
	token, err := lockToken()
	if err != nil {
		return nil, false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	redisKey := l.prefix + ":dispatch:lock:" + key
	ok, err := l.client.SetNX(ctx, redisKey, token, ttl).Result()
	if err != nil || !ok {
		return nil, ok, err
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			const releaseScript = `
				if redis.call("GET", KEYS[1]) == ARGV[1] then
					return redis.call("DEL", KEYS[1])
				end
				return 0
			`
			_ = l.client.Eval(ctx, releaseScript, []string{redisKey}, token).Err()
		})
	}, true, nil
}

func lockToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
