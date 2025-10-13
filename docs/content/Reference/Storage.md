---
sidebar_position: 4
tags: [developer-guide, storage]
---

# Storage

COO-LLM uses pluggable storage backends for runtime metrics and caching. The system supports Redis, in-memory, HTTP API, and file-based storage.

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
    type: "redis"  # redis, memory, http, file
    addr: "localhost:6379"
    password: ""
    api_key: ""
```

### Cache Configuration

```yaml
policy:
  cache:
    enabled: true
    ttl_seconds: 10
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

## Metrics Usage

Metrics are used for:

- **Rate Limiting:** Check req/min and tokens/min limits
- **Load Balancing:** Select least-loaded keys
- **Monitoring:** Track performance and errors
- **Caching:** Response deduplication

## Best Practices

### Production Setup

- Use Redis for production deployments
- Set appropriate TTL values (default 60s)
- Monitor Redis memory usage
- Implement Redis persistence (RDB/AOF)

### Development Setup

- Use in-memory storage for quick testing
- Switch to Redis when testing load balancing
- Check logs for storage errors

## Troubleshooting

### Common Issues

**Redis connection failed:**
```bash
# Test connection
redis-cli -h localhost -p 6379 ping

# Check Redis logs
docker logs redis-container
```

**Metrics not updating:**
- Verify storage backend configuration
- Check for storage errors in logs
- Ensure proper permissions

**High memory usage:**
- Monitor Redis memory with `INFO memory`
- Adjust TTL values if needed
- Implement key expiration

### Debug Commands

```bash
# View all usage keys
redis-cli keys "usage:*"

# Get specific metric
redis-cli get "usage:openai:key1:req"

# Check TTL
redis-cli ttl "usage:openai:key1:req"
```