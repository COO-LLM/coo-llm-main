---
sidebar_position: 5
tags: [reference, storage, http, api]
---

# HTTP API Storage

REST API-based storage backend for integrating with external storage services.

## Configuration

```yaml
storage:
  runtime:
    type: "http"
    addr: "https://api.example.com/storage"
    api_key: "${STORAGE_API_KEY}"
    timeout: "30s"  # Optional, default 30s
```

## Features

- **REST API Integration**: HTTP-based storage operations
- **Flexible Endpoints**: Custom API endpoints for different operations
- **Authentication**: API key and custom header support
- **Timeout Handling**: Configurable request timeouts
- **Error Handling**: HTTP status code mapping
- **JSON Serialization**: Standard JSON data format
- **Custom Headers**: Support for additional authentication headers

## API Endpoints

The HTTP storage expects REST endpoints:

```
GET    /usage/{provider}/{key_id}/{metric}     # Get usage value
POST   /usage/{provider}/{key_id}/{metric}     # Set usage value  
POST   /usage/{provider}/{key_id}/{metric}/inc # Increment usage
GET    /cache/{key}                           # Get cached value
POST   /cache/{key}                           # Set cached value with TTL
```

## Request/Response Format

**Usage Operations:**
```json
// GET /usage/openai/key1/req
{
  "value": 45.0
}

// POST /usage/openai/key1/req
{
  "value": 46.0
}
```

**Cache Operations:**
```json
// GET /cache/mykey
{
  "value": "cached_data"
}

// POST /cache/mykey
{
  "value": "new_data",
  "ttl_seconds": 300
}
```

## Implementation Details

- **HTTP Client**: Configurable timeout and retry logic
- **Authentication**: Bearer token or custom headers
- **Serialization**: JSON encoding/decoding
- **Error Mapping**: HTTP status to storage errors
- **Connection Pooling**: Efficient HTTP connection management
