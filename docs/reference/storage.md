# Storage

---
sidebar_position: 4
tags: [developer-guide, storage]
---

TruckLLM uses multiple storage backends for different purposes: runtime metrics, configuration, and logging. The system supports pluggable storage providers for maximum flexibility.

## Storage Types

### Runtime Storage

Stores transient data like usage metrics, request counts, and performance statistics.

**Requirements:**
- High write throughput
- Time-based expiration (TTL)
- Atomic operations
- Fast reads

**Supported Backends:**
- Redis (default)
- HTTP API
- In-memory (development)

### Configuration Storage

Persists system configuration and supports dynamic reloading.

**Requirements:**
- Persistent storage
- Atomic updates
- Versioning support
- Backup/restore

**Supported Backends:**
- File system (YAML)
- HTTP API

### Logging Storage

Handles log aggregation and forwarding.

**Requirements:**
- High write throughput
- Structured data support
- Filtering and search
- External system integration

**Supported Backends:**
- File system
- HTTP webhooks
- Prometheus pushgateway

## Redis Storage (Default)

### Configuration

```yaml
storage:
  runtime:
    type: "redis"
    addr: "localhost:6379"
    password: ""
    db: 0
```

### Data Structure

**Usage Metrics:**
```
Key: usage:{provider}:{key_id}:{metric}
Value: float64
TTL: 60 seconds

Examples:
usage:openai:key1:req → 45.0
usage:openai:key1:tokens → 1200.0
usage:openai:key1:errors → 2.0
usage:openai:key1:latency → 1250.0
```

**Atomic Operations:**
- `INCRBYFLOAT` for usage increments
- `EXPIRE` for TTL management
- `MGET` for bulk metric retrieval

### Lua Scripts

For complex operations:

```lua
-- Atomic usage update with TTL
local key = KEYS[1]
local value = ARGV[1]
local ttl = ARGV[2]

redis.call('INCRBYFLOAT', key, value)
redis.call('EXPIRE', key, ttl)
```

## HTTP API Storage

### Configuration

```yaml
storage:
  runtime:
    type: "http"
    addr: "https://api.example.com/storage"
    api_key: "your-api-key"
```

### API Endpoints

**Get Usage:**
```
GET /usage/{provider}/{keyId}/{metric}
Authorization: Bearer {api_key}

Response: "123.45"
```

**Set Usage:**
```
PUT /usage/{provider}/{keyId}/{metric}
Authorization: Bearer {api_key}
Content-Type: application/json

{"value": 123.45}
```

**Increment Usage:**
```
POST /usage/{provider}/{keyId}/{metric}/increment
Authorization: Bearer {api_key}
Content-Type: application/json

{"delta": 10.0}
```

**Load Config:**
```
GET /config
Authorization: Bearer {api_key}

Response: {...config...}
```

**Save Config:**
```
PUT /config
Authorization: Bearer {api_key}
Content-Type: application/json

{...config...}
```

### Implementation

```go
type HTTPStore struct {
    endpoint string
    apiKey   string
}

func (h *HTTPStore) GetUsage(provider, keyID, metric string) (float64, error) {
    url := fmt.Sprintf("%s/usage/%s/%s/%s", h.endpoint, provider, keyID, metric)
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+h.apiKey)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return 0, err
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    return strconv.ParseFloat(string(body), 64)
}
```

## File Storage

### Configuration

```yaml
storage:
  config:
    type: "file"
    path: "./configs/config.yaml"
```

### Usage

- Configuration persistence
- Development/testing environments
- Backup storage

### Limitations

- No concurrent access control
- No atomic operations
- File system dependent

## Storage Interface

All storage backends implement common interfaces:

```go
type RuntimeStore interface {
    GetUsage(provider, keyID, metric string) (float64, error)
    SetUsage(provider, keyID, metric string, value float64) error
    IncrementUsage(provider, keyID, metric string, delta float64) error
}

type ConfigStore interface {
    LoadConfig() (*config.Config, error)
    SaveConfig(cfg *config.Config) error
}
```

## Metrics Collection

### Usage Tracking

**Automatic Updates:**
- Request count incremented on each API call
- Token usage extracted from provider responses
- Error counts tracked for reliability monitoring
- Latency measured and stored

**Sliding Windows:**
- Metrics stored with TTL (60 seconds)
- Aggregated over time windows
- Used for rate limiting and load balancing

### Data Retention

**Runtime Data:**
- Short-term: 60 seconds (active metrics)
- Medium-term: 1 hour (performance analysis)
- Long-term: External monitoring systems

**Configuration:**
- Persistent storage
- Version history (optional)
- Backup procedures

## Performance Considerations

### Redis Optimization

- Connection pooling
- Pipelining for bulk operations
- Memory-efficient data structures
- Cluster support for scaling

### HTTP API Optimization

- Connection reuse
- Request batching
- Circuit breaker pattern
- Retry logic with backoff

### Caching Strategy

- In-memory LRU cache for hot metrics
- Write-through caching for consistency
- Cache invalidation on updates

## Reliability

### Failure Handling

**Storage Unavailable:**
- Graceful degradation to in-memory storage
- Alert generation
- Automatic retry with exponential backoff

**Data Consistency:**
- Atomic operations where possible
- Transaction support in Redis
- Conflict resolution for concurrent updates

### Backup & Recovery

**Configuration Backup:**
- Automatic snapshots
- Version control integration
- Disaster recovery procedures

**Metrics Backup:**
- Export to external systems
- Historical data archiving
- Compliance requirements

## Monitoring

### Storage Metrics

```prometheus
# Storage operation latency
storage_operation_duration_seconds{operation="get",backend="redis"} 0.001

# Storage errors
storage_errors_total{backend="redis",operation="set"} 5

# Connection pool stats
storage_connections_active{backend="redis"} 10
storage_connections_idle{backend="redis"} 5
```

### Health Checks

**Redis Health:**
```go
func (r *RedisStore) Health() error {
    return r.client.Ping(context.Background()).Err()
}
```

**HTTP Health:**
```go
func (h *HTTPStore) Health() error {
    resp, err := http.Get(h.endpoint + "/health")
    if err != nil {
        return err
    }
    return resp.Body.Close()
}
```

## Security

### Authentication

- API key authentication for HTTP storage
- Redis password protection
- TLS encryption for network communication

### Authorization

- Scoped access control
- Provider/key-level permissions
- Audit logging for access

### Data Protection

- Encryption at rest
- Secure key management
- Compliance with data protection regulations

## Best Practices

### Configuration

- Use Redis for production deployments
- Implement HTTP API for centralized storage
- Regular backup of configuration

### Monitoring

- Set up alerts for storage failures
- Monitor storage performance metrics
- Implement health check endpoints

### Scaling

- Use Redis Cluster for high availability
- Implement read replicas for metrics
- Consider sharding for large deployments

## Troubleshooting

### Common Issues

**Redis Connection Failed:**
- Check network connectivity
- Verify authentication credentials
- Monitor connection pool usage

**HTTP API Timeout:**
- Check endpoint availability
- Review network configuration
- Implement retry logic

**Data Inconsistency:**
- Verify atomic operation usage
- Check for concurrent access issues
- Implement conflict resolution

### Debug Commands

```bash
# Redis connection test
redis-cli -h localhost -p 6379 ping

# Check usage metrics
redis-cli keys "usage:*"

# HTTP API test
curl -H "Authorization: Bearer key" http://api.example.com/health
```

## Future Enhancements

- **Database Support:** PostgreSQL, MySQL backends
- **Distributed Storage:** etcd, Consul integration
- **Time-Series Databases:** InfluxDB, Prometheus storage
- **Caching Layer:** Redis caching for configuration
- **Multi-Region:** Cross-region replication