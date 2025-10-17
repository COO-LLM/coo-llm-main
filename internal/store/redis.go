package store

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog"
)

type RedisStore struct {
	client *redis.Client
	logger zerolog.Logger
}

func NewRedisStore(addr, password string, logger zerolog.Logger) *RedisStore {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	return &RedisStore{client: rdb, logger: logger}
}

func (r *RedisStore) GetUsage(provider, keyID, metric string) (float64, error) {
	key := fmt.Sprintf("usage:%s:%s:%s", provider, keyID, metric)
	val, err := r.client.Get(context.Background(), key).Float64()
	if err == redis.Nil {
		r.logger.Debug().Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", 0).Msg("store operation - key not found")
		return 0, nil
	}
	if err != nil {
		r.logger.Error().Err(err).Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Msg("store operation failed")
		return 0, err
	}
	r.logger.Debug().Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", val).Msg("store operation")
	return val, nil
}

func (r *RedisStore) SetUsage(provider, keyID, metric string, value float64) error {
	key := fmt.Sprintf("usage:%s:%s:%s", provider, keyID, metric)
	err := r.client.Set(context.Background(), key, value, time.Minute).Err()
	if err != nil {
		r.logger.Error().Err(err).Str("operation", "SetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", value).Msg("store operation failed")
		return err
	}
	r.logger.Debug().Str("operation", "SetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", value).Msg("store operation")
	return nil
}

func (r *RedisStore) IncrementUsage(provider, keyID, metric string, delta float64) error {
	// Store with timestamp for sliding window
	timestampKey := fmt.Sprintf("usage:%s:%s:%s:%d", provider, keyID, metric, time.Now().Unix())
	err := r.client.Set(context.Background(), timestampKey, delta, time.Hour).Err()
	if err != nil {
		r.logger.Error().Err(err).Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("store operation failed - timestamp key")
		return err
	}

	// Also maintain total counter
	totalKey := fmt.Sprintf("usage:%s:%s:%s", provider, keyID, metric)
	err = r.client.IncrByFloat(context.Background(), totalKey, delta).Err()
	if err != nil {
		r.logger.Error().Err(err).Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("store operation failed - total counter")
		return err
	}
	r.logger.Debug().Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("store operation")
	return nil
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

func (r *RedisStore) StoreMetric(name string, value float64, tags map[string]string, timestamp int64) error {
	ctx := context.Background()
	key := fmt.Sprintf("metrics:%s", name)
	// Store as sorted set with timestamp as score, value as member
	member := fmt.Sprintf("%f", value)
	return r.client.ZAdd(ctx, key, &redis.Z{Score: float64(timestamp), Member: member}).Err()
}

func (r *RedisStore) GetMetrics(name string, tags map[string]string, start, end int64) ([]MetricPoint, error) {
	ctx := context.Background()
	key := fmt.Sprintf("metrics:%s", name)
	results, err := r.client.ZRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", start),
		Max: fmt.Sprintf("%d", end),
	}).Result()
	if err != nil {
		return nil, err
	}
	var points []MetricPoint
	for _, z := range results {
		var value float64
		fmt.Sscanf(z.Member.(string), "%f", &value)
		points = append(points, MetricPoint{
			Value:     value,
			Timestamp: int64(z.Score),
			Tags:      make(map[string]string),
		})
	}
	return points, nil
}
