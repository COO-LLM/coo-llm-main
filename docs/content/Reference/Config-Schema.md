---
sidebar_position: 1
tags: [reference, config]
---

# Configuration Schema

Complete YAML schema reference for COO-LLM configuration.

## Full Schema

```yaml
version: "1.0"  # Config version

server:
  listen: ":2906"  # Server listen address
  admin_api_key: "${ADMIN_API_KEY}"  # Admin API key
  cors:
    enabled: true  # Enable CORS
    allowed_origins: ["*"]  # Allowed origins or ["*"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]  # Allowed HTTP methods
    allowed_headers: ["*"]  # Allowed headers or ["*"]
    allow_credentials: true  # Allow credentials
    max_age: 86400  # Preflight cache duration in seconds
  webui:
    enabled: true  # Enable web UI
    admin_id: "admin"  # Web UI admin username
    admin_password: "password"  # Web UI admin password
    web_ui_path: "/path/to/custom/ui"  # Optional custom UI path

logging:
  file:
    enabled: true  # Enable file logging
    path: "./logs/llm.log"  # Log file path
    max_size_mb: 100  # Max file size in MB
    max_backups: 5  # Max backup files
  prometheus:
    enabled: true  # Enable Prometheus metrics
    endpoint: "/metrics"  # Metrics endpoint
  providers: []  # Log provider configs

storage:
  config:
    type: "file"  # Config storage type
    path: "./data/config.json"  # Config file path
  runtime:
    type: "sql"  # Runtime storage type
    addr: "localhost:6379"  # Redis/InfluxDB address
    password: "${REDIS_PASSWORD}"  # Redis password
    api_key: "${INFLUX_TOKEN}"  # InfluxDB token
    database: "${INFLUX_BUCKET}"  # Database/bucket name

llm_providers:
  - id: "openai"  # Unique provider ID
    name: "OpenAI"  # Display name
    type: "openai"  # Provider type
    api_keys: ["${OPENAI_KEY_1}", "${OPENAI_KEY_2}"]  # API keys as ENV vars (resolved at runtime)
    base_url: "https://api.openai.com"  # API base URL
    model: "gpt-4o"  # Default model
    pricing:
      input_token_cost: 0.002  # Cost per input token
      output_token_cost: 0.01  # Cost per output token
    limits:
      req_per_min: 200  # Requests per minute per key
      tokens_per_min: 100000  # Tokens per minute per key
      max_tokens: 4096  # Max tokens per request
      session_limit: 6000000  # Session token limit
      session_type: "1h"  # Session window

api_keys:
  - id: "client-001"  # Optional unique identifier
    key: "${API_KEY}"  # Client API key
    allowed_providers: ["*"]  # Allowed providers or ["*"]
    description: "Client description"

model_aliases: {}  # Model alias mappings (deprecated)

policy:
  strategy: "hybrid"  # Legacy field
  algorithm: "round_robin"  # Selection algorithm
  priority: "balanced"  # Preset weights
  hybrid_weights:
    token_ratio: 0.2  # Token usage weight
    req_ratio: 0.2  # Request count weight
    error_score: 0.2  # Error rate weight
    latency: 0.2  # Latency weight
    cost_ratio: 0.2  # Cost weight
  retry:
    max_attempts: 3  # Max retry attempts
    timeout: "30s"  # Request timeout
    interval: "1s"  # Retry interval
   cache:
     enabled: true  # Enable response caching
     ttl_seconds: 10  # Cache TTL
```

## Provider Configuration

COO-LLM supports multiple LLM providers with unified interface:

### Supported Providers

| Provider | Type | Features |
|----------|------|----------|
| OpenAI | `openai` | GPT models, embeddings, vision |
| Anthropic Claude | `claude` | Claude models, advanced reasoning |
| Google Gemini | `gemini` | Gemini models, multimodal |
| Grok | `grok` | xAI Grok models |
| Fireworks | `fireworks` | Open-source models |
| Together AI | `together` | Various open-source models |
| Replicate | `replicate` | Model deployment platform |
| Hugging Face | `huggingface` | Open-source models |
| Mistral | `mistral` | Mistral models |
| Cohere | `cohere` | Command models, embeddings |
| Voyage AI | `voyage` | Embedding models |
| OpenRouter | `openrouter` | Unified API access |

### Provider Configuration

```yaml
llm_providers:
  - id: "openai-prod"  # Unique identifier
    type: "openai"     # Provider type
    api_keys: ["${OPENAI_KEY1}", "${OPENAI_KEY2}"]  # ENV vars resolved at runtime
    base_url: "https://api.openai.com/v1"  # Optional custom endpoint
    model: "gpt-4o"    # Default model
    pricing:
      input_token_cost: 0.0025   # Cost per input token
      output_token_cost: 0.01    # Cost per output token
    limits:
      req_per_min: 200           # Rate limits
      tokens_per_min: 100000
      max_tokens: 4096
      session_limit: 1000000     # Token session limits
      session_type: "1d"         # Window for limits
```

### Key Features

- **Load Balancing**: Automatic key rotation and provider selection
- **Rate Limiting**: Per-key and global limits with session tracking
- **Cost Tracking**: Real-time cost calculation and monitoring
- **Error Handling**: Retry logic with fallback providers
- **Caching**: Response caching for performance
- **Security**: API keys resolved from ENV, never stored

See individual provider docs in `Developer-Guide/Providers/` for setup details.

## Configuration Management

COO-LLM supports dynamic configuration with the following features:

- **ENV Resolution**: API keys can be specified as `${ENV_VAR}` and are resolved from environment variables at runtime
- **Store Sync**: Public configuration is saved to store on startup and can be updated via admin API
- **Security**: Sensitive data (API keys) are never stored in the config store, only ENV placeholders
- **Instance Sync**: Multiple instances share public config from store, with local ENV for secrets

### Admin API for Config Updates

Use the admin API to update configuration dynamically:

```bash
# Get current config
curl -H "Authorization: Bearer admin-key" http://localhost:2906/api/admin/v1/config

# Update config
curl -X POST -H "Authorization: Bearer admin-key" \
  -H "Content-Type: application/json" \
  http://localhost:2906/api/admin/v1/config \
  -d '{"api_keys": [{"key": "new-key", "allowed_providers": ["openai"]}]}'
```

## Field Types & Validation

### Server

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `listen` | string | Yes | `:2906` | Valid port format |
| `admin_api_key` | string | Yes | - | Non-empty |
| `cors.enabled` | bool | No | `true` | - |
| `cors.allowed_origins` | []string | No | `["*"]` | Valid origins or `["*"]` |
| `cors.allowed_methods` | []string | No | `["GET", "POST", "PUT", "DELETE", "OPTIONS"]` | Valid HTTP methods |
| `cors.allowed_headers` | []string | No | `["*"]` | Valid headers or `["*"]` |
| `cors.allow_credentials` | bool | No | `true` | - |
| `cors.max_age` | int | No | `86400` | >= 0 |
| `webui.enabled` | bool | No | `true` | - |
| `webui.admin_id` | string | No | `admin` | Non-empty |
| `webui.admin_password` | string | No | `password` | Non-empty |
| `webui.web_ui_path` | string | No | - | Valid path if set |

### Logging

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `file.enabled` | bool | No | `true` | - |
| `file.path` | string | No | `./logs/llm.log` | Valid path |
| `file.max_size_mb` | int | No | `100` | > 0 |
| `file.max_backups` | int | No | `5` | > 0 |
| `prometheus.enabled` | bool | No | `true` | - |
| `prometheus.endpoint` | string | No | `/metrics` | Valid path |

### Storage

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `config.type` | string | No | `file` | `file`, `http` |
| `config.path` | string | No | `./data/config.json` | Valid path |
| `runtime.type` | string | No | `sql` | `redis`, `http`, `influxdb`, `mongodb`, `sql`, `dynamodb` |
| `runtime.addr` | string | No | - | Valid URL |
| `runtime.password` | string | No | - | - |
| `runtime.api_key` | string | No | - | - |
| `runtime.database` | string | No | - | - |

### LLM Providers

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `id` | string | Yes | - | Unique, non-empty |
| `name` | string | No | - | - |
| `type` | string | Yes | - | `openai`, `gemini`, `claude`, `custom` |
| `api_keys` | []string | Yes | - | At least 1 key |
| `base_url` | string | No | - | Valid URL |
| `model` | string | Yes | - | Non-empty |
| `pricing.input_token_cost` | float64 | No | `0` | >= 0 |
| `pricing.output_token_cost` | float64 | No | `0` | >= 0 |
| `limits.req_per_min` | int | No | `0` | >= 0 |
| `limits.tokens_per_min` | int | No | `0` | >= 0 |
| `limits.max_tokens` | int | No | `0` | >= 0 |
| `limits.session_limit` | int | No | `0` | >= 0 |
| `limits.session_type` | string | No | `1h` | Valid duration |

### API Keys

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `id` | string | No | - | Unique identifier for the API key |
| `key` | string | Yes | - | Unique, non-empty |
| `allowed_providers` | []string | No | `["*"]` | Valid provider IDs or `["*"]` |
| `description` | string | No | - | - |

### Policy

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `algorithm` | string | No | `round_robin` | `round_robin`, `least_loaded`, `hybrid` |
| `priority` | string | No | `balanced` | `balanced`, `cost`, `req`, `token` |
| `hybrid_weights.*` | float64 | No | - | 0.0-1.0 |
| `retry.max_attempts` | int | No | `3` | > 0 |
| `retry.timeout` | duration | No | `30s` | Valid duration |
| `retry.interval` | duration | No | `1s` | Valid duration |
| `cache.enabled` | bool | No | `true` | - |
| `cache.ttl_seconds` | int64 | No | `10` | > 0 |

## Environment Variables

All string fields support environment variable substitution:

```yaml
server:
  admin_api_key: "${ADMIN_API_KEY}"
  listen: ":${PORT}"

llm_providers:
  - api_keys: ["${OPENAI_KEY_1}", "${OPENAI_KEY_2}"]
```

Variables are expanded at config load time. Use `export VAR=value` before running.

## Validation Rules

- **Required fields**: `version`, `server.listen`, `server.admin_api_key`, `llm_providers[].id`, `llm_providers[].type`, `llm_providers[].api_keys`, `llm_providers[].model`
- **CORS**: When enabled, validates allowed origins, methods, and headers
- **Unique constraints**: Provider IDs, API key values
- **Type validation**: Numbers must be numeric, URLs must be valid
- **Cross-references**: `api_keys[].allowed_providers` must reference valid provider IDs

## Migration Guide

### From Legacy Providers

Old format:
```yaml
providers:
  - id: "openai"
    keys:
      - secret: "sk-key"
        pricing: {...}
```

New format:
```yaml
llm_providers:
  - id: "openai"
    api_keys: ["sk-key"]
    pricing: {...}
```

### Deprecated Fields

- `model_aliases`: Use provider:model syntax directly
- `policy.strategy`: Use `policy.algorithm`
- Legacy `providers` array: Migrated to `llm_providers`