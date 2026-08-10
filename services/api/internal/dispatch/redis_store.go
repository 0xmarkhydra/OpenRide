package dispatch

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisOfferStore struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

func NewRedisOfferStore(client *redis.Client, prefix string) *RedisOfferStore {
	if prefix == "" {
		prefix = "flashx"
	}
	return &RedisOfferStore{client: client, prefix: prefix, ttl: 24 * time.Hour}
}

func (s *RedisOfferStore) Put(offer Offer) error {
	payload, err := json.Marshal(offer)
	if err != nil {
		return err
	}
	ctx, cancel := dispatchRedisContext()
	defer cancel()

	created, err := s.client.SetNX(ctx, s.offerKey(offer.ID), payload, s.ttl).Result()
	if err != nil {
		return err
	}
	if !created {
		return ErrOfferUnavailable
	}
	pipe := s.client.TxPipeline()
	pipe.SAdd(ctx, s.tripKey(offer.TripID), offer.ID)
	pipe.Expire(ctx, s.tripKey(offer.TripID), s.ttl)
	pipe.SAdd(ctx, s.driverKey(offer.DriverID), offer.ID)
	pipe.Expire(ctx, s.driverKey(offer.DriverID), s.ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		_ = s.client.Del(ctx, s.offerKey(offer.ID)).Err()
		return err
	}
	return nil
}

func (s *RedisOfferStore) Get(id string) (Offer, error) {
	ctx, cancel := dispatchRedisContext()
	defer cancel()
	payload, err := s.client.Get(ctx, s.offerKey(id)).Bytes()
	if errors.Is(err, redis.Nil) {
		return Offer{}, ErrOfferNotFound
	}
	if err != nil {
		return Offer{}, err
	}
	var offer Offer
	if err := json.Unmarshal(payload, &offer); err != nil {
		return Offer{}, err
	}
	return offer, nil
}

func (s *RedisOfferStore) Save(offer Offer) error {
	payload, err := json.Marshal(offer)
	if err != nil {
		return err
	}
	ctx, cancel := dispatchRedisContext()
	defer cancel()
	exists, err := s.client.Exists(ctx, s.offerKey(offer.ID)).Result()
	if err != nil {
		return err
	}
	if exists == 0 {
		return ErrOfferNotFound
	}
	return s.client.Set(ctx, s.offerKey(offer.ID), payload, s.ttl).Err()
}

func (s *RedisOfferStore) ListByTrip(tripID string) ([]Offer, error) {
	return s.list(s.tripKey(tripID))
}

func (s *RedisOfferStore) ListByDriver(driverID string) ([]Offer, error) {
	return s.list(s.driverKey(driverID))
}

func (s *RedisOfferStore) list(indexKey string) ([]Offer, error) {
	ctx, cancel := dispatchRedisContext()
	defer cancel()
	ids, err := s.client.SMembers(ctx, indexKey).Result()
	if err != nil {
		return nil, err
	}
	items := make([]Offer, 0, len(ids))
	for _, id := range ids {
		payload, getErr := s.client.Get(ctx, s.offerKey(id)).Bytes()
		if errors.Is(getErr, redis.Nil) {
			_ = s.client.SRem(ctx, indexKey, id).Err()
			continue
		}
		if getErr != nil {
			return nil, getErr
		}
		var offer Offer
		if err := json.Unmarshal(payload, &offer); err != nil {
			return nil, err
		}
		items = append(items, offer)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items, nil
}

func (s *RedisOfferStore) offerKey(id string) string {
	return s.prefix + ":dispatch:offer:" + id
}

func (s *RedisOfferStore) tripKey(id string) string {
	return s.prefix + ":dispatch:trip:" + id
}

func (s *RedisOfferStore) driverKey(id string) string {
	return s.prefix + ":dispatch:driver:" + id
}

func dispatchRedisContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}
