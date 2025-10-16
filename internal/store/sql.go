package store

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

type SQLStore struct {
	db     *sql.DB
	logger zerolog.Logger
}

func NewSQLStore(connStr string, logger zerolog.Logger) (*SQLStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create tables if not exist
	if err := createTables(db); err != nil {
		return nil, err
	}

	return &SQLStore{db: db, logger: logger}, nil
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS usage_metrics (
			id SERIAL PRIMARY KEY,
			provider VARCHAR(50) NOT NULL,
			key_id VARCHAR(100) NOT NULL,
			metric VARCHAR(50) NOT NULL,
			value DOUBLE PRECISION NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(provider, key_id, metric)
		)`,
		`CREATE TABLE IF NOT EXISTS usage_history (
			id SERIAL PRIMARY KEY,
			provider VARCHAR(50) NOT NULL,
			key_id VARCHAR(100) NOT NULL,
			metric VARCHAR(50) NOT NULL,
			delta DOUBLE PRECISION NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS cache (
			key VARCHAR(255) PRIMARY KEY,
			value TEXT NOT NULL,
			expiry TIMESTAMP WITH TIME ZONE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_metrics_provider_key_metric ON usage_metrics(provider, key_id, metric)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_history_timestamp ON usage_history(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_history_provider_key_metric ON usage_history(provider, key_id, metric)`,
		`CREATE INDEX IF NOT EXISTS idx_cache_expiry ON cache(expiry)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLStore) GetUsage(provider, keyID, metric string) (float64, error) {
	var value float64
	err := s.db.QueryRow(
		"SELECT value FROM usage_metrics WHERE provider = $1 AND key_id = $2 AND metric = $3",
		provider, keyID, metric,
	).Scan(&value)

	if err == sql.ErrNoRows {
		s.logger.Debug().Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", 0).Msg("store operation - no rows")
		return 0, nil
	}
	if err != nil {
		s.logger.Error().Err(err).Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Msg("store operation failed")
		return 0, err
	}

	s.logger.Debug().Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", value).Msg("store operation")
	return value, nil
}

func (s *SQLStore) SetUsage(provider, keyID, metric string, value float64) error {
	_, err := s.db.Exec(
		`INSERT INTO usage_metrics (provider, key_id, metric, value) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (provider, key_id, metric) DO UPDATE SET value = EXCLUDED.value`,
		provider, keyID, metric, value,
	)
	if err != nil {
		s.logger.Error().Err(err).Str("operation", "SetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", value).Msg("store operation failed")
		return err
	}
	s.logger.Debug().Str("operation", "SetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", value).Msg("store operation")
	return nil
}

func (s *SQLStore) IncrementUsage(provider, keyID, metric string, delta float64) error {
	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error().Err(err).Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("transaction begin failed")
		return err
	}
	defer tx.Rollback()

	// Insert into history
	_, err = tx.Exec(
		"INSERT INTO usage_history (provider, key_id, metric, delta) VALUES ($1, $2, $3, $4)",
		provider, keyID, metric, delta,
	)
	if err != nil {
		s.logger.Error().Err(err).Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("insert history failed")
		return err
	}

	// Update or insert total
	_, err = tx.Exec(
		`INSERT INTO usage_metrics (provider, key_id, metric, value) VALUES ($1, $2, $3, $4)
		 ON CONFLICT (provider, key_id, metric) DO UPDATE SET value = usage_metrics.value + EXCLUDED.value`,
		provider, keyID, metric, delta,
	)
	if err != nil {
		s.logger.Error().Err(err).Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("update total failed")
		return err
	}

	if err = tx.Commit(); err != nil {
		s.logger.Error().Err(err).Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("transaction commit failed")
		return err
	}

	s.logger.Debug().Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Msg("store operation")
	return nil
}

func (s *SQLStore) GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error) {
	var total float64
	err := s.db.QueryRow(
		"SELECT COALESCE(SUM(delta), 0) FROM usage_history WHERE provider = $1 AND key_id = $2 AND metric = $3 AND timestamp > NOW() - INTERVAL '1 second' * $4",
		provider, keyID, metric, windowSeconds,
	).Scan(&total)

	if err != nil {
		s.logger.Error().Err(err).Str("operation", "GetUsageInWindow").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Int64("windowSeconds", windowSeconds).Msg("store operation failed")
		return 0, err
	}

	s.logger.Debug().Str("operation", "GetUsageInWindow").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Int64("windowSeconds", windowSeconds).Float64("total", total).Msg("store operation")
	return total, nil
}

func (s *SQLStore) SetCache(key, value string, ttlSeconds int64) error {
	expiry := time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	_, err := s.db.Exec(
		`INSERT INTO cache (key, value, expiry) VALUES ($1, $2, $3)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, expiry = EXCLUDED.expiry`,
		key, value, expiry,
	)
	if err != nil {
		s.logger.Error().Err(err).Str("operation", "SetCache").Str("key", key).Int64("ttlSeconds", ttlSeconds).Msg("store operation failed")
		return err
	}
	s.logger.Debug().Str("operation", "SetCache").Str("key", key).Int64("ttlSeconds", ttlSeconds).Msg("store operation")
	return nil
}

func (s *SQLStore) GetCache(key string) (string, error) {
	var value string
	var expiry pq.NullTime
	err := s.db.QueryRow(
		"SELECT value, expiry FROM cache WHERE key = $1 AND (expiry IS NULL OR expiry > NOW())",
		key,
	).Scan(&value, &expiry)

	if err == sql.ErrNoRows {
		s.logger.Debug().Str("operation", "GetCache").Str("key", key).Msg("store operation - cache miss")
		return "", nil
	}
	if err != nil {
		s.logger.Error().Err(err).Str("operation", "GetCache").Str("key", key).Msg("store operation failed")
		return "", err
	}

	s.logger.Debug().Str("operation", "GetCache").Str("key", key).Msg("store operation - cache hit")
	return value, nil
}
