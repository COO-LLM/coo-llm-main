---
sidebar_position: 2
tags: [reference, storage, mongodb]
---

# MongoDB Storage

Document-based NoSQL database for flexible and scalable COO-LLM data storage.

## Configuration

```yaml
storage:
  runtime:
    type: "mongodb"
    addr: "mongodb://localhost:27017"
    database: "coo_llm"
    username: "${MONGO_USER}"  # Optional
    password: "${MONGO_PASS}"  # Optional
```

## Features

- **Document Model**: JSON-like documents with dynamic schemas
- **Flexible Schema**: No fixed table structures
- **Aggregation Pipeline**: Advanced data processing and analytics
- **Indexing**: Automatic and custom indexes for performance
- **Replica Sets**: High availability and data redundancy
- **Sharding**: Horizontal scaling across multiple servers
- **GridFS**: Large file storage support

## Data Structure

```javascript
// Usage metrics collection
{
    "_id": ObjectId("..."),
    "provider": "openai",
    "key_id": "key1", 
    "metric": "req",
    "value": 45.0,
    "updated_at": ISODate("2024-01-01T00:00:00Z")
}

// Usage history collection
{
    "_id": ObjectId("..."),
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

## Implementation Details

- **Connection Pooling**: Efficient connection management
- **BSON Serialization**: Native MongoDB document format
- **Index Management**: Automatic index creation and optimization
- **Aggregation Queries**: Complex analytics with pipeline stages
- **Change Streams**: Real-time data change notifications
