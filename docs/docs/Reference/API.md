---
sidebar_position: 2
tags: [developer-guide, api]
---

# API Reference

COO-LLM provides OpenAI-compatible REST APIs for LLM interactions, plus administrative endpoints for management.

## Authentication

All API requests require authentication via Bearer token:

```
Authorization: Bearer <api_key>
```

API keys are configured in the providers section and mapped to specific keys.

## OpenAI-Compatible Endpoints

### POST /v1/chat/completions

Generate chat completions using available models. This is the primary endpoint implemented in COO-LLM.

**Request Body:**
```json
{
  "model": "gpt-4o",
  "messages": [
    {
      "role": "user",
      "content": "Hello, how are you?"
    }
  ],
  "max_tokens": 100
}
```

**Response:**
```json
{
  "id": "chatcmpl-1234567890",
  "object": "chat.completion",
  "created": 1699123456,
  "model": "gpt-4o",
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
- `model` (string, required): Model alias from configuration (e.g., "gpt-4o", "gemini-1.5-pro")
- `messages` (array, required): Chat messages with role/content format
- `max_tokens` (integer, optional): Maximum tokens to generate (default: 1000)
- Additional parameters are passed through to the provider

**Features:**
- ✅ Conversation history support
- ✅ Model alias resolution
- ✅ Automatic provider selection and load balancing
- ✅ Rate limiting and retry logic
- ✅ Response caching (if enabled)
- ✅ Usage tracking and metrics

### GET /v1/models

List available models based on configured model aliases.

**Response:**
```json
{
  "object": "list",
  "data": [
    {
      "id": "gpt-4o",
      "object": "model",
      "created": 1699123456,
      "owned_by": "coo-llm"
    },
    {
      "id": "gemini-1.5-pro",
      "object": "model",
      "created": 1699123456,
      "owned_by": "coo-llm"
    }
  ]
}
```

**Note:** Models are listed based on `model_aliases` configuration, not actual provider models.

## Admin API Endpoints

**Note:** Admin API endpoints are not yet implemented in the current version. The following are planned for future releases:

### Planned Admin Endpoints

- `GET /admin/v1/config` - Get current configuration
- `POST /admin/v1/config` - Update configuration
- `POST /admin/v1/config/validate` - Validate configuration
- `POST /admin/v1/reload` - Reload configuration from file
- `GET /admin/v1/providers` - Get provider status and metrics
- `GET /admin/v1/logs` - Get recent log entries

### Current Admin Access

Currently, configuration management is done via:
- Configuration file reloading (restart required)
- Direct file editing
- Environment variable changes

Admin functionality will be added in future versions.

## Metrics Endpoint

### GET /metrics

Prometheus metrics endpoint (enabled when `logging.prometheus.enabled: true`).

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

**Available Metrics:**
- Request counts by provider/model
- Request duration histograms
- Error rates
- Token usage tracking
- Active connections

**Configuration:**
```yaml
logging:
  prometheus:
    enabled: true
    endpoint: "/metrics"
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

COO-LLM implements rate limiting based on configured limits:

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

# Point to COO-LLM instead of OpenAI
client = openai.OpenAI(
    api_key="dummy-key",  # COO-LLM ignores this, uses config-based auth
    base_url="http://localhost:8080/v1"
)

response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello!"}]
)

print(response.choices[0].message.content)
```

### cURL Examples

```bash
# Chat completion with API key auth
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer test-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello"}]
  }'

# List available models
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer test-key"

# Prometheus metrics
curl http://localhost:8080/metrics
```

## Authentication

COO-LLM uses API key authentication via the `Authorization: Bearer <key>` header. API keys are configured in the `api_keys` section and map to allowed providers.

```yaml
api_keys:
  - key: "client-a-key"
    allowed_providers: ["openai-prod"]  # Limited access
  - key: "premium-key"
    allowed_providers: ["*"]  # Full access
```

## SDK Compatibility

COO-LLM is compatible with:

- OpenAI Python SDK (`openai>=1.0`)
- OpenAI Node.js SDK
- Any HTTP client following OpenAI Chat Completions API format

Simply change the `base_url` to point to your COO-LLM instance and use any API key from your configuration.