# API Reference

---
sidebar_position: 1
---

TruckLLM provides OpenAI-compatible REST APIs for LLM interactions, plus administrative endpoints for management.

## Authentication

All API requests require authentication via Bearer token:

```
Authorization: Bearer <api_key>
```

API keys are configured in the providers section and mapped to specific keys.

## OpenAI-Compatible Endpoints

### POST /v1/chat/completions

Generate chat completions using available models.

**Request Body:**
```json
{
  "model": "gpt-4",
  "messages": [
    {
      "role": "user",
      "content": "Hello, how are you?"
    }
  ],
  "max_tokens": 100,
  "temperature": 0.7
}
```

**Response:**
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! I'm doing well, thank you for asking."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 13,
    "completion_tokens": 7,
    "total_tokens": 20
  }
}
```

**Parameters:**
- `model` (string, required): Model alias from configuration
- `messages` (array, required): Chat messages
- `max_tokens` (integer): Maximum tokens to generate
- `temperature` (float): Sampling temperature (0.0 to 2.0)
- `top_p` (float): Nucleus sampling parameter
- `stream` (boolean): Enable streaming responses

### POST /v1/completions

Generate text completions.

**Request Body:**
```json
{
  "model": "gpt-3.5-turbo-instruct",
  "prompt": "The quick brown fox",
  "max_tokens": 50
}
```

**Response:**
```json
{
  "id": "cmpl-123",
  "object": "text_completion",
  "created": 1677652288,
  "model": "gpt-3.5-turbo-instruct",
  "choices": [
    {
      "text": " jumps over the lazy dog.",
      "index": 0,
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 5,
    "completion_tokens": 6,
    "total_tokens": 11
  }
}
```

### POST /v1/embeddings

Generate embeddings for text.

**Request Body:**
```json
{
  "model": "text-embedding-ada-002",
  "input": "The food was delicious and the service was excellent."
}
```

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "object": "embedding",
      "embedding": [0.1, 0.2, ...],
      "index": 0
    }
  ],
  "model": "text-embedding-ada-002",
  "usage": {
    "prompt_tokens": 8,
    "total_tokens": 8
  }
}
```

### GET /v1/models

List available models.

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4",
      "object": "model",
      "created": 1687882411,
      "owned_by": "openai"
    },
    {
      "id": "gemini-1.5-pro",
      "object": "model",
      "created": 1687882411,
      "owned_by": "gemini"
    }
  ]
}
```

## Admin API Endpoints

Admin endpoints require the `admin_api_key` configured in server settings.

### GET /admin/v1/config

Get current configuration.

**Response:**
```json
{
  "version": "1.0",
  "server": {...},
  "providers": [...],
  ...
}
```

### POST /admin/v1/config

Update configuration.

**Request Body:** Full configuration object

**Response:** Success message

### POST /admin/v1/config/validate

Validate configuration without applying.

**Request Body:** Configuration to validate

**Response:**
```json
{
  "valid": true
}
```

### POST /admin/v1/reload

Reload configuration from file.

**Response:** Success message

### GET /admin/v1/providers

Get provider status and metrics.

**Response:**
```json
{
  "providers": [
    {
      "id": "openai",
      "name": "OpenAI",
      "status": "healthy",
      "keys": [
        {
          "id": "key1",
          "usage_req": 45,
          "usage_tokens": 1200,
          "errors": 2
        }
      ]
    }
  ]
}
```

### GET /admin/v1/logs

Get recent log entries.

**Query Parameters:**
- `limit` (integer): Number of entries (default: 100)
- `provider` (string): Filter by provider
- `level` (string): Filter by log level

**Response:**
```json
{
  "logs": [
    {
      "timestamp": "2024-01-01T12:00:00Z",
      "level": "info",
      "provider": "openai",
      "model": "gpt-4",
      "latency_ms": 1200,
      "status": 200,
      "tokens": 150
    }
  ]
}
```

## Metrics Endpoint

### GET /metrics

Prometheus metrics endpoint (if enabled).

**Response:** Prometheus format metrics

```
# HELP llm_requests_total Total number of LLM requests
# TYPE llm_requests_total counter
llm_requests_total{provider="openai",model="gpt-4"} 1250

# HELP llm_request_duration_seconds Request duration in seconds
# TYPE llm_request_duration_seconds histogram
llm_request_duration_seconds_bucket{provider="openai",le="0.1"} 1200
...
```

## Error Responses

All endpoints return standard HTTP status codes:

- `200`: Success
- `400`: Bad Request (invalid parameters)
- `401`: Unauthorized (invalid API key)
- `403`: Forbidden (rate limited or quota exceeded)
- `404`: Not Found (invalid endpoint or model)
- `429`: Too Many Requests (rate limited)
- `500`: Internal Server Error
- `502`: Bad Gateway (provider error)
- `503`: Service Unavailable (provider down)

Error response format:
```json
{
  "error": {
    "message": "Invalid model specified",
    "type": "invalid_request_error",
    "code": 400
  }
}
```

## Rate Limiting

TruckLLM implements rate limiting based on configured limits:

- Per-key request limits
- Per-key token limits
- Global rate limits

Rate limited requests return `429` status with retry information.

## Streaming

Streaming responses are supported for chat completions:

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your-key" \
  -d '{"model": "gpt-4", "messages": [{"role": "user", "content": "Tell me a story"}], "stream": true}'
```

Response is Server-Sent Events format.

## Examples

### Python Client

```python
import openai

# Point to TruckLLM instead of OpenAI
client = openai.OpenAI(
    api_key="your-truckllm-key",
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="gpt-4",
    messages=[{"role": "user", "content": "Hello!"}]
)

print(response.choices[0].message.content)
```

### cURL Examples

```bash
# Chat completion
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello"}]
  }'

# List models
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer your-key"

# Get config (admin)
curl http://localhost:8080/admin/v1/config \
  -H "Authorization: Bearer your-admin-key"
```

## SDK Compatibility

TruckLLM is compatible with:

- OpenAI Python SDK
- OpenAI Node.js SDK
- OpenAI Go SDK
- Any HTTP client following OpenAI API format

Simply change the `base_url` to point to your TruckLLM instance.