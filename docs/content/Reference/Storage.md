---
sidebar_position: 4
tags: [developer-guide, storage]
---

# Storage

COO-LLM uses pluggable storage backends for runtime metrics and caching. The system supports Redis, in-memory, HTTP API, file-based, SQL databases, MongoDB, and DynamoDB storage.

## Storage Interface

All storage backends implement the `RuntimeStore` interface from `internal/store/interface.go`:

```go
type RuntimeStore interface {
    GetUsage(provider, keyID, metric string) (float64, error)
    SetUsage(provider, keyID, metric string, value float64) error
    IncrementUsage(provider, keyID, metric string, delta float64) error
    GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error)
    SetCache(key, value string, ttlSeconds int64) error
    GetCache(key string) (string, error)
}
```

## Supported Backends

### SQL Database Storage (PostgreSQL)

**Configuration:**
```yaml
storage:
  runtime:
    type: "sql"
    addr: "postgresql://user:password@localhost/dbname?sslmode=disable"
```

**Features:**
- Full SQL database support with PostgreSQL
- Persistent storage with ACID transactions
- Advanced querying capabilities
- Time-window analytics support
- Automatic table creation and indexing

**Data Structure:**
```sql
-- Usage metrics table
CREATE TABLE usage_metrics (
    provider VARCHAR(50) NOT NULL,
    key_id VARCHAR(100) NOT NULL,
    metric VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    UNIQUE(provider, key_id, metric)
);

-- Usage history table for time-window queries
CREATE TABLE usage_history (
    provider VARCHAR(50) NOT NULL,
    key_id VARCHAR(100) NOT NULL,
    metric VARCHAR(50) NOT NULL,
    delta DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Cache table
CREATE TABLE cache (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    expiry TIMESTAMP WITH TIME ZONE
);
```

### MongoDB Storage

**Configuration:**
```yaml
storage:
  runtime:
    type: "mongodb"
    addr: "mongodb://localhost:27017"
    database: "coo_llm"
```

**Features:**
- NoSQL document database support
- Flexible schema design
- High performance for read/write operations
- Aggregation pipeline for analytics
- Automatic index creation

**Data Structure:**
```javascript
// Usage metrics collection
{
    "_id": ObjectId,
    "provider": "openai",
    "key_id": "key1",
    "metric": "req",
    "value": 45.0
}

// Usage history collection
{
    "_id": ObjectId,
    "provider": "openai",
    "key_id": "key1",
    "metric": "req",
    "delta": 1.0,
    "timestamp": ISODate("2024-01-01T00:00:00Z")
}

// Cache collection
{
    "_id": "cache_key",
    "value": "cached_response",
    "expiry": ISODate("2024-01-01T00:00:10Z")
}
```

### DynamoDB Storage (AWS)

**Configuration:**
```yaml
storage:
  runtime:
    type: "dynamodb"
    addr: "us-east-1"  # AWS region
    table_usage: "coo_llm_usage"
    table_cache: "coo_llm_cache"
    table_history: "coo_llm_history"
```

**Features:**
- AWS managed NoSQL database
- Auto-scaling and high availability
- Pay-per-request pricing
- Global tables for multi-region
- Time-window queries with history table

**Data Structure:**
```
Usage Table:
PK: USAGE#{provider}#{key_id}
SK: {metric}
Attributes: value (Number)

History Table:
PK: HISTORY#{provider}#{key_id}#{metric}
SK: {timestamp}
Attributes: delta (Number), timestamp (Number)

Cache Table:
PK: CACHE#{key}
SK: DATA
Attributes: value (String), expiry (Number)
```

### Redis Storage (Production)

**Configuration:**
```yaml
storage:
  runtime:
    type: "redis"
    addr: "localhost:6379"
    password: "${REDIS_PASSWORD}"
```

**Features:**
- Persistent storage for metrics
- High performance with connection pooling
- TTL support for automatic cleanup
- Used for production deployments

**Data Structure:**
```
Key: usage:{provider}:{key_id}:{metric}
Value: float64
TTL: 60 seconds

Examples:
usage:openai:key1:req → 45.0
usage:openai:key1:tokens → 1200.0
usage:openai:key1:errors → 2.0
```

### In-Memory Storage (Development)

**Configuration:**
```yaml
storage:
  runtime:
    type: "memory"
```

**Features:**
- Fast in-memory storage
- No persistence (lost on restart)
- Used for development and testing

### HTTP API Storage

**Configuration:**
```yaml
storage:
  runtime:
    type: "http"
    addr: "https://api.example.com/storage"
    api_key: "${STORAGE_API_KEY}"
```

**Features:**
- Remote storage via HTTP API
- Useful for centralized metrics storage

### File Storage

**Configuration:**
```yaml
storage:
  runtime:
    type: "file"
    path: "./storage/data.json"
```

**Features:**
- Simple file-based storage
- Not recommended for production

## Usage Metrics

The system tracks the following metrics per provider/key combination:

- `req`: Number of requests
- `input_tokens`: Input tokens used
- `output_tokens`: Output tokens generated
- `tokens`: Total tokens used
- `errors`: Number of failed requests
- `latency`: Average response latency in milliseconds

## Caching

Response caching is supported when enabled:

```yaml
policy:
  cache:
    enabled: true
    ttl_seconds: 10
```

**Cache Keys:** Normalized text prompts
**Storage:** Same backend as runtime metrics
**TTL:** Configurable expiration time

## Configuration

### Runtime Storage

```yaml
storage:
  runtime:
    type: "redis"  # redis, memory, http, file, sql, mongodb, dynamodb
    addr: "localhost:6379"  # Connection string or endpoint
    password: ""            # Redis password
    api_key: ""             # HTTP API key
    database: "coo_llm"     # Database name for MongoDB
    table_usage: "coo_llm_usage"     # DynamoDB table names
    table_cache: "coo_llm_cache"
    table_history: "coo_llm_history"
```

### Cache Configuration

```yaml
policy:
  cache:
    enabled: true
    ttl_seconds: 10
```

## Logging and Monitoring

All storage backends include comprehensive logging for database operations:

- **Debug logs**: All Get/Set/Increment operations with parameters and results
- **Error logs**: Failed operations with error details
- **Performance monitoring**: Operation timing and success rates

**Log Levels:**
- `DEBUG`: Successful operations with context
- `ERROR`: Failed operations with error messages
- `INFO`: Connection status and initialization

**Example logs:**
```
DEBUG store operation operation=GetUsage provider=openai keyID=key1 metric=req value=45
DEBUG store operation operation=IncrementUsage provider=openai keyID=key1 metric=tokens delta=150
ERROR store operation failed operation=SetUsage provider=openai keyID=key1 metric=req error="connection timeout"
```

## Implementation Details

### Redis Backend

**File:** `internal/store/redis.go`

**Features:**
- Uses go-redis client
- Automatic TTL management
- Atomic increment operations
- Connection pooling

### Memory Backend

**File:** `internal/store/memory.go`

**Features:**
- Thread-safe with sync.RWMutex
- In-memory map storage
- No persistence
- Fast for development

### HTTP Backend

**File:** `internal/store/http.go`

**Features:**
- REST API integration
- Bearer token authentication
- JSON request/response format

### File Backend

**File:** `internal/store/file.go`

**Features:**
- JSON file storage
- Simple persistence
- Not concurrent-safe

### SQL Backend

**File:** `internal/store/sql.go`

**Features:**
- PostgreSQL driver with `lib/pq`
- Automatic schema migration
- Transaction support for consistency
- Indexed queries for performance

### MongoDB Backend

**File:** `internal/store/mongodb.go`

**Features:**
- Official MongoDB Go driver
- Connection pooling and monitoring
- Aggregation framework for analytics
- Automatic index management

### DynamoDB Backend

**File:** `internal/store/dynamodb.go`

**Features:**
- AWS SDK v2 for DynamoDB
- Multi-table architecture for efficiency
- Conditional updates and atomic operations
- Time-window queries via history table

## Metrics Usage

Metrics are used for:

- **Rate Limiting:** Check req/min and tokens/min limits
- **Load Balancing:** Select least-loaded keys
- **Monitoring:** Track performance and errors
- **Caching:** Response deduplication

## Best Practices

### Production Setup

**Redis:**
- Use Redis for production deployments
- Set appropriate TTL values (default 60s)
- Monitor Redis memory usage
- Implement Redis persistence (RDB/AOF)

**SQL Databases:**
- Use PostgreSQL for relational workloads
- Enable connection pooling
- Monitor query performance and indexes
- Regular VACUUM for maintenance

**MongoDB:**
- Use MongoDB for high-throughput scenarios
- Configure replica sets for HA
- Monitor oplog and disk usage
- Use appropriate read preferences

**DynamoDB:**
- Use DynamoDB for AWS-native deployments
- Monitor read/write capacity and costs
- Design partition keys for even distribution
- Use DynamoDB Streams for cross-region replication

### Development Setup

- Use in-memory storage for quick testing
- Switch to Redis when testing load balancing
- Check logs for storage errors

## Troubleshooting

### Common Issues

**Database Connection Failed:**
```bash
# Redis
redis-cli -h localhost -p 6379 ping

# PostgreSQL
psql -h localhost -U user -d dbname -c "SELECT 1"

# MongoDB
mongosh --eval "db.runCommand({ping: 1})"

# DynamoDB (via AWS CLI)
aws dynamodb list-tables --region us-east-1
```

**Metrics not updating:**
- Verify storage backend configuration
- Check for storage errors in logs
- Ensure proper permissions and credentials
- Validate connection strings

**High memory/disk usage:**
- Monitor database metrics
- Adjust TTL values if needed
- Implement data cleanup policies
- Check for memory leaks in application

**Slow queries:**
- Add appropriate indexes
- Monitor query execution plans
- Consider data partitioning
- Use connection pooling

### Debug Commands

**Redis:**
```bash
# View all usage keys
redis-cli keys "usage:*"

# Get specific metric
redis-cli get "usage:openai:key1:req"

# Check TTL
redis-cli ttl "usage:openai:key1:req"
```

**PostgreSQL:**
```sql
-- View usage metrics
SELECT * FROM usage_metrics WHERE provider = 'openai';

-- Check recent history
SELECT * FROM usage_history
WHERE timestamp > NOW() - INTERVAL '1 hour'
ORDER BY timestamp DESC;

-- Monitor table sizes
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables WHERE tablename LIKE 'usage%';
```

**MongoDB:**
```javascript
// View usage metrics
db.usage_metrics.find({provider: "openai"})

// Check recent history
db.usage_history.find({
  timestamp: {$gt: new Date(Date.now() - 3600000)}
}).sort({timestamp: -1})

// Monitor collection stats
db.usage_metrics.stats()
```

**DynamoDB:**
```bash
# List tables
aws dynamodb list-tables --region us-east-1

# Scan usage table
aws dynamodb scan --table-name coo_llm_usage --region us-east-1

# Query history
aws dynamodb query \
  --table-name coo_llm_history \
  --key-condition-expression "pk = :pk" \
  --expression-attribute-values '{":pk":{"S":"HISTORY#openai#key1#req"}}' \
  --region us-east-1
```