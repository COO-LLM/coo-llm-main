---
sidebar_position: 1
tags: [reference, storage, redis]
---

# Redis Storage

High-performance in-memory data structure store for COO-LLM runtime metrics and caching.

## Configuration

```yaml
storage:
  runtime:
    type: "redis"
    addr: "localhost:6379"
    password: "${REDIS_PASSWORD}"  # Optional
    database: 0  # Optional, default 0
```

## Features

- **High Performance**: In-memory storage with sub-millisecond access
- **Persistence**: RDB snapshots and AOF for durability
- **Data Structures**: Hashes, sorted sets, lists for complex data
- **Pub/Sub**: Real-time messaging capabilities
- **Clustering**: Horizontal scaling with Redis Cluster
- **TTL Support**: Automatic key expiration
- **Atomic Operations**: Thread-safe increments and updates

## Data Structure

Redis uses key-value pairs with structured naming:

```
# Usage metrics (hash)
usage:{provider}:{key_id}:{metric} -> hash of values

# Usage history (sorted set with timestamps)
usage_history:{provider}:{key_id}:{metric} -> score:timestamp, member:delta

# Cache (string with TTL)
cache:{key} -> value (with expiry)
```

## Implementation Details

- **Connection Pooling**: Automatic connection management
- **Serialization**: JSON encoding for complex data
- **Error Handling**: Automatic reconnection on failures
- **Pipeline Support**: Batch operations for performance
