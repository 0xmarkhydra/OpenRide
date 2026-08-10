package drivers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLocationIndex struct {
	client *redis.Client
	prefix string
}

// NewRedisLocationIndex accepts an optional namespace so integration tests,
// staging and production can share a Redis server without key collisions.
func NewRedisLocationIndex(client *redis.Client, prefix ...string) *RedisLocationIndex {
	namespace := "flashx"
	if len(prefix) > 0 && prefix[0] != "" {
		namespace = prefix[0]
	}
	return &RedisLocationIndex{client: client, prefix: namespace}
}

type redisLocationRecord struct {
	ServiceType string   `json:"service_type"`
	Location    Location `json:"location"`
	Eligible    bool     `json:"eligible"`
}

func (r *RedisLocationIndex) Upsert(driverID, serviceType string, location Location, eligible bool) error {
	ctx, cancel := redisContext()
	defer cancel()
	record := redisLocationRecord{ServiceType: serviceType, Location: location, Eligible: eligible}
	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, r.locationKey(driverID), payload, 5*time.Minute)
	if eligible {
		pipe.GeoAdd(ctx, r.geoKey(serviceType), &redis.GeoLocation{Name: driverID, Longitude: location.Lng, Latitude: location.Lat})
	} else {
		pipe.ZRem(ctx, r.geoKey(serviceType), driverID)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisLocationIndex) SetEligible(driverID, serviceType string, eligible bool) error {
	ctx, cancel := redisContext()
	defer cancel()
	payload, err := r.client.Get(ctx, r.locationKey(driverID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return ErrLocationNotFound
	}
	if err != nil {
		return err
	}
	var record redisLocationRecord
	if err := json.Unmarshal(payload, &record); err != nil {
		return err
	}
	oldServiceType := record.ServiceType
	record.ServiceType = serviceType
	record.Eligible = eligible
	updated, err := json.Marshal(record)
	if err != nil {
		return err
	}
	pipe := r.client.TxPipeline()
	pipe.Set(ctx, r.locationKey(driverID), updated, 5*time.Minute)
	if oldServiceType != "" && oldServiceType != serviceType {
		pipe.ZRem(ctx, r.geoKey(oldServiceType), driverID)
	}
	if eligible {
		pipe.GeoAdd(ctx, r.geoKey(serviceType), &redis.GeoLocation{Name: driverID, Longitude: record.Location.Lng, Latitude: record.Location.Lat})
	} else {
		pipe.ZRem(ctx, r.geoKey(serviceType), driverID)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (r *RedisLocationIndex) Get(driverID string) (Location, error) {
	ctx, cancel := redisContext()
	defer cancel()
	payload, err := r.client.Get(ctx, r.locationKey(driverID)).Bytes()
	if errors.Is(err, redis.Nil) {
		return Location{}, ErrLocationNotFound
	}
	if err != nil {
		return Location{}, err
	}
	var record redisLocationRecord
	if err := json.Unmarshal(payload, &record); err != nil {
		return Location{}, err
	}
	return record.Location, nil
}

func (r *RedisLocationIndex) Nearby(serviceType string, center Location, radiusM float64, limit int) ([]NearbyLocation, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if radiusM <= 0 {
		radiusM = 5000
	}
	ctx, cancel := redisContext()
	defer cancel()
	results, err := r.client.GeoSearchLocation(ctx, r.geoKey(serviceType), &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  center.Lng,
			Latitude:   center.Lat,
			Radius:     radiusM,
			RadiusUnit: "m",
			Sort:       "ASC",
			Count:      limit,
		},
		WithCoord: true,
		WithDist:  true,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("redis geosearch: %w", err)
	}
	items := make([]NearbyLocation, 0, len(results))
	for _, result := range results {
		location, err := r.Get(result.Name)
		if errors.Is(err, ErrLocationNotFound) {
			_ = r.client.ZRem(ctx, r.geoKey(serviceType), result.Name).Err()
			continue
		}
		if err != nil {
			return nil, err
		}
		items = append(items, NearbyLocation{DriverID: result.Name, Location: location, DistanceM: result.Dist})
	}
	return items, nil
}

func (r *RedisLocationIndex) geoKey(serviceType string) string {
	return r.prefix + ":geo:drivers:" + serviceType
}

func (r *RedisLocationIndex) locationKey(driverID string) string {
	return r.prefix + ":driver:location:" + driverID
}

func redisContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}
