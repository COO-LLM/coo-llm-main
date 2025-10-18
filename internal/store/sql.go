package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
)

type SQLStore struct {
	db     *sql.DB
	logger zerolog.Logger
	dbType string // "postgres" or "sqlite"
}

func NewSQLStore(connStr string, logger zerolog.Logger) (*SQLStore, error) {
	var db *sql.DB
	var dbType string
	var err error

	// Detect database type from connection string
	if strings.Contains(connStr, "sqlite") || strings.HasSuffix(connStr, ".db") || strings.HasSuffix(connStr, ".sqlite") || strings.HasPrefix(connStr, "./") || strings.HasPrefix(connStr, "/") {
		db, err = sql.Open("sqlite3", connStr)
		dbType = "sqlite"
	} else {
		db, err = sql.Open("postgres", connStr)
		dbType = "postgres"
	}

	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create tables if not exist
	if err := createTables(db, dbType); err != nil {
		return nil, err
	}

	return &SQLStore{db: db, logger: logger, dbType: dbType}, nil
}

// placeholder returns the appropriate placeholder for the database type
func (s *SQLStore) placeholder(n int) string {
	if s.dbType == "sqlite" {
		return "?"
	}
	return fmt.Sprintf("$%d", n)
}

func createTables(db *sql.DB, dbType string) error {

	var queries []string
	if dbType == "sqlite" {
		queries = []string{
			`CREATE TABLE IF NOT EXISTS usage_metrics (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				provider TEXT NOT NULL,
				key_id TEXT NOT NULL,
				metric TEXT NOT NULL,
				value REAL NOT NULL,
				timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(provider, key_id, metric)
			)`,
			`CREATE TABLE IF NOT EXISTS usage_history (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				provider TEXT NOT NULL,
				key_id TEXT NOT NULL,
				metric TEXT NOT NULL,
				delta REAL NOT NULL,
				timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
			)`,
			`CREATE TABLE IF NOT EXISTS cache (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL,
				expiry DATETIME
			)`,
			`CREATE INDEX IF NOT EXISTS idx_usage_metrics_provider_key_metric ON usage_metrics(provider, key_id, metric)`,
			`CREATE INDEX IF NOT EXISTS idx_usage_history_timestamp ON usage_history(timestamp)`,
			`CREATE INDEX IF NOT EXISTS idx_usage_history_provider_key_metric ON usage_history(provider, key_id, metric)`,
			`CREATE INDEX IF NOT EXISTS idx_cache_expiry ON cache(expiry)`,
		}
	} else {
		// PostgreSQL queries
		queries = []string{
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
	var query string
	if s.dbType == "sqlite" {
		query = "SELECT value FROM usage_metrics WHERE provider = ? AND key_id = ? AND metric = ?"
	} else {
		query = "SELECT value FROM usage_metrics WHERE provider = $1 AND key_id = $2 AND metric = $3"
	}
	err := s.db.QueryRow(query, provider, keyID, metric).Scan(&value)

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

func (s *SQLStore) StoreMetric(name string, value float64, tags map[string]string, timestamp int64) error {
	tagsJSON, _ := json.Marshal(tags)
	_, err := s.db.Exec(
		"INSERT INTO metrics (name, value, tags, timestamp) VALUES ($1, $2, $3, $4)",
		name, value, string(tagsJSON), timestamp,
	)
	return err
}

func (s *SQLStore) GetMetrics(name string, tags map[string]string, start, end int64) ([]MetricPoint, error) {
	rows, err := s.db.Query(
		"SELECT value, timestamp FROM metrics WHERE name = $1 AND timestamp >= $2 AND timestamp <= $3",
		name, start, end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []MetricPoint
	for rows.Next() {
		var point MetricPoint
		err := rows.Scan(&point.Value, &point.Timestamp)
		if err != nil {
			return nil, err
		}
		point.Tags = make(map[string]string)
		points = append(points, point)
	}
	return points, nil
}
