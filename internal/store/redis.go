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
	key := fmt.Sprintf("usage:%s:%s:%s", provider, keyID, metric)
	return r.client.IncrByFloat(context.Background(), key, delta).Err()
}
