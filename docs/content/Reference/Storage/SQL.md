---
sidebar_position: 8
tags: [reference, storage, sql, postgresql]
---

# SQL Database Storage

Full-featured SQL database storage with ACID transactions and advanced querying.

## Configuration

```yaml
storage:
  runtime:
    type: "sql"
    addr: "postgresql://user:password@localhost/dbname?sslmode=disable"
    max_open_conns: 25    # Optional, default 25
    max_idle_conns: 5     # Optional, default 5
    conn_max_lifetime: "1h" # Optional
```

## Features

- **ACID Transactions**: Atomic, Consistent, Isolated, Durable operations
- **SQL Queries**: Full SQL power for complex analytics
- **Indexing**: Automatic and custom database indexes
- **Joins & Aggregations**: Advanced data relationships and analytics
- **Backup/Restore**: Standard database backup procedures
- **High Availability**: Replication and failover support
- **Scalability**: Vertical and horizontal scaling options

## Data Structure

```sql
-- Usage metrics table
CREATE TABLE usage_metrics (
    provider VARCHAR(50) NOT NULL,
    key_id VARCHAR(100) NOT NULL,
    metric VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (provider, key_id, metric)
);

-- Usage history for time-window queries
CREATE TABLE usage_history (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(50) NOT NULL,
    key_id VARCHAR(100) NOT NULL,
    metric VARCHAR(50) NOT NULL,
    delta DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Cache table with TTL
CREATE TABLE cache (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    expiry TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_usage_history_provider_key_metric_time 
ON usage_history(provider, key_id, metric, timestamp);

CREATE INDEX idx_usage_history_time ON usage_history(timestamp);
CREATE INDEX idx_cache_expiry ON cache(expiry) WHERE expiry IS NOT NULL;
```

## Implementation Details

- **Connection Pooling**: Efficient database connection management
- **Prepared Statements**: SQL injection prevention and performance
- **Transaction Management**: Atomic operations across multiple tables
- **Migration Support**: Automatic schema creation and updates
- **Query Optimization**: Efficient SQL queries with proper indexing
- **Error Handling**: Database-specific error mapping
