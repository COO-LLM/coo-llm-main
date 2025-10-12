package store

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore(addr, password string) *RedisStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &RedisStore{client: rdb}
}

func (r *RedisStore) GetUsage(provider, keyID, metric string) (float64, error) {
	key := fmt.Sprintf("usage:%s:%s:%s", provider, keyID, metric)
	val, err := r.client.Get(context.Background(), key).Float64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

func (r *RedisStore) SetUsage(provider, keyID, metric string, value float64) error {
	key := fmt.Sprintf("usage:%s:%s:%s", provider, keyID, metric)
	return r.client.Set(context.Background(), key, value, time.Minute).Err()
}

func (r *RedisStore) IncrementUsage(provider, keyID, metric string, delta float64) error {
	// Store with timestamp for sliding window
	timestampKey := fmt.Sprintf("usage:%s:%s:%s:%d", provider, keyID, metric, time.Now().Unix())
	err := r.client.Set(context.Background(), timestampKey, delta, time.Hour).Err()
	if err != nil {
		return err
	}

	// Also maintain total counter
	totalKey := fmt.Sprintf("usage:%s:%s:%s", provider, keyID, metric)
	return r.client.IncrByFloat(context.Background(), totalKey, delta).Err()
}

func (r *RedisStore) GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error) {
	ctx := context.Background()
	now := time.Now().Unix()
	start := now - windowSeconds

	// Find all timestamp keys in window
	pattern := fmt.Sprintf("usage:%s:%s:%s:*", provider, keyID, metric)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, err
	}

	total := 0.0
	for _, key := range keys {
		// Extract timestamp from key
		var timestamp int64
		fmt.Sscanf(key, fmt.Sprintf("usage:%s:%s:%s:%%d", provider, keyID, metric), &timestamp)

		if timestamp >= start && timestamp <= now {
			val, err := r.client.Get(ctx, key).Float64()
			if err == nil {
				total += val
			}
		}
	}

	return total, nil
}

func (r *RedisStore) SetCache(key, value string, ttlSeconds int64) error {
	cacheKey := fmt.Sprintf("cache:%s", key)
	return r.client.Set(context.Background(), cacheKey, value, time.Duration(ttlSeconds)*time.Second).Err()
}

func (r *RedisStore) GetCache(key string) (string, error) {
	cacheKey := fmt.Sprintf("cache:%s", key)
	val, err := r.client.Get(context.Background(), cacheKey).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}
