---
sidebar_position: 2
tags: [reference, api, admin]
---

# Admin API Reference

Administrative endpoints for configuration and monitoring (require admin authentication).

## Authentication

Admin endpoints require the `admin_api_key` configured in `server.admin_api_key`:

```
Authorization: Bearer <admin_api_key>
```

## Security Features

### Rate Limiting

Admin endpoints are protected by rate limiting:
- **Limit**: 100 requests per minute per IP address
- **Response**: HTTP 429 (Too Many Requests) when exceeded

### Audit Logging

All admin API requests are logged with:
- Request method and path
- Client IP address
- User agent
- Response time
- Success/failure status

### CORS Support

Admin API endpoints support Cross-Origin Resource Sharing (CORS) for web applications:

- **Preflight Requests**: OPTIONS requests are automatically handled
- **Headers**: Standard CORS headers are included in all responses
- **Configuration**: CORS behavior is controlled by `server.cors` settings
- **Credentials**: Supports credentialed requests when enabled

**Example CORS Headers:**
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Content-Type, Authorization
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 86400
```

## Configuration Management

### GET /admin/v1/config

Get current configuration.

**Response:**
```json
{
  "version": "1.0",
  "server": {
    "listen": ":2906",
    "admin_api_key": "****"
  },
  "llm_providers": [...],
  "api_keys": [
    {
      "id": "client-001",
      "key": "sk-****",
      "allowed_providers": ["openai"],
      "description": "Production client"
    }
  ],
  "policy": {...}
}
```

### POST /admin/v1/config

Update configuration (hot-reload).

**Request Body:**
```json
{
  "llm_providers": [
    {
      "id": "openai",
      "api_keys": ["sk-new-key"]
    }
  ]
}
```

**Response:**
```json
{
  "status": "configuration updated",
  "reloaded": true
}
```

### POST /admin/v1/config/validate

Validate configuration without applying.

**Request Body:** Full config JSON

**Response:**
```json
{
  "valid": true,
  "errors": []
}
```

### PUT /admin/v1/config/policy

Update the load balancing policy.

**Request Body:**
```json
{
  "algorithm": "round_robin",
  "priority": "latency",
  "cache": {
    "enabled": true,
    "ttl_seconds": 60
  }
}
```

**Response:**
```json
{
  "message": "Policy updated successfully",
  "policy": {
    "algorithm": "round_robin",
    "priority": "latency",
    "cache": {
      "enabled": true,
      "ttl_seconds": 60
    }
  }
}
```

## Client Management

### POST /admin/v1/clients

Create a new API client.

**Request Body:**
```json
{
  "client_id": "client-123",
  "api_key": "sk-client-key",
  "description": "Production client for app",
  "allowed_providers": ["openai", "anthropic"]
}
```

**Response:**
```json
{
  "status": "created",
  "client_id": "client-123"
}
```

### GET /admin/v1/clients/list

List all API clients.

**Response:**
```json
{
  "clients": [
    {
      "id": "client-123",
      "api_key": "sk-****",
      "description": "Production client",
      "allowed_providers": ["openai", "anthropic"],
      "created_at": 1700000000,
      "last_used": 1700001000
    }
  ]
}
```

### GET /admin/v1/clients/\{client_id\}

Get details for a specific client.

**Response:**
```json
{
  "id": "client-123",
  "api_key": "sk-client-key",
  "description": "Production client",
  "allowed_providers": ["openai", "anthropic"],
  "created_at": 1700000000,
  "last_used": 1700001000
}
```

### PUT /admin/v1/clients/\{client_id\}

Update client configuration.

**Request Body:**
```json
{
  "description": "Updated description",
  "allowed_providers": ["openai"]
}
```

**Response:**
```json
{
  "status": "updated",
  "client_id": "client-123"
}
```

### DELETE /admin/v1/clients/\{client_id\}

Delete an API client.

**Response:**
```json
{
  "status": "deleted",
  "client_id": "client-123"
}
```



## Monitoring & Metrics

### GET /admin/v1/metrics

Get historical metrics.

**Query Parameters:**
- `name`: Metric name (latency, tokens, cost)
- `start`: Start timestamp (Unix seconds)
- `end`: End timestamp (Unix seconds)

**Response:**
```json
{
  "name": "latency",
  "start": 1700000000,
  "end": 1700003600,
  "points": [
    {"timestamp": 1700000100, "value": 150.5},
    {"timestamp": 1700000200, "value": 120.2}
  ]
}
```



### GET /admin/v1/clients

Get client API key usage.

**Response:**
```json
{
  "clients": [
    {
      "key": "client-a",
      "requests": 500,
      "tokens": 10000,
      "last_used": 1700000000
    }
  ]
}
```

### GET /admin/v1/stats

Get overall system statistics.

**Response:**
```json
{
  "uptime": 3600,
  "total_requests": 1000,
  "total_tokens": 50000,
  "total_cost": 25.50,
  "providers": {
    "openai": {"requests": 600, "cost": 15.00},
    "gemini": {"requests": 400, "cost": 10.50}
  }
}
```



## Enhanced Metrics

### GET /admin/v1/metrics/clients/\{client_id\}

Get detailed metrics for a specific client.

**Query Parameters:**
- `start`: Start timestamp (Unix seconds)
- `end`: End timestamp (Unix seconds)

**Response:**
```json
{
  "client_id": "client-123",
  "total_requests": 1000,
  "total_tokens": 50000,
  "total_cost": 25.50,
  "success_rate": 0.98,
  "avg_latency": 150.5,
  "last_request_time": 1700001000
}
```

### GET /admin/v1/metrics/providers/\{provider_id\}

Get detailed metrics for a specific provider.

**Query Parameters:**
- `start`: Start timestamp (Unix seconds)
- `end`: End timestamp (Unix seconds)

**Response:**
```json
{
  "provider_id": "openai",
  "total_requests": 500,
  "total_tokens": 25000,
  "total_cost": 12.50,
  "success_rate": 0.99,
  "avg_latency": 120.3,
  "error_count": 5
}
```

### GET /admin/v1/metrics/global

Get global system metrics.

**Query Parameters:**
- `start`: Start timestamp (Unix seconds)
- `end`: End timestamp (Unix seconds)

**Response:**
```json
{
  "total_clients": 5,
  "total_providers": 3,
  "total_requests": 2500,
  "total_tokens": 125000,
  "total_cost": 62.50,
  "overall_success_rate": 0.97,
  "avg_latency": 135.2
}
```



## Web UI Authentication

### POST /admin/login

Authenticate for web UI access.

**Request Body:**
```json
{
  "id": "admin",
  "password": "password"
}
```

**Response:**
```json
{
  "token": "webui-admin-1700000000",
  "expires": 1700003600
}
```

## Error Responses

All admin API errors follow this format:

```json
{
  "error": "invalid configuration",
  "details": "missing required field: admin_api_key"
}
```

## Rate Limiting

Admin endpoints are protected by rate limiting to prevent abuse:
- **Limit**: 100 requests per minute per IP address
- **Response**: HTTP 429 (Too Many Requests) when exceeded
- **Scope**: Applies to all admin API endpoints

All admin requests are logged for audit purposes including:
- Request details (method, path, IP, user agent)
- Response time and status
- Authentication attempts