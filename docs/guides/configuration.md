# Configuration

---
sidebar_position: 2
tags: [user-guide, configuration]
---

TruckLLM uses YAML configuration files for all settings. The configuration is hierarchical and supports environment variable substitution, validation, and hot-reload.

## Configuration File Structure

```yaml
version: "1.0"

server:
  listen: ":8080"
  admin_api_key: "your-admin-key"

logging:
  file:
    enabled: true
    path: "./logs/llm.log"
    max_size_mb: 100
    max_backups: 5
  prometheus:
    enabled: true
    endpoint: "/metrics"
  providers:
    - name: "webhook"
      type: "http"
      endpoint: "https://logs.example.com/ingest"
      batch:
        enabled: true
        size: 50
        interval_seconds: 10

storage:
  config:
    type: "file"
    path: "./configs/config.yaml"
  runtime:
    type: "redis"
    addr: "localhost:6379"
    password: ""
    api_key: ""

providers:
  - id: "openai"
    name: "OpenAI"
    base_url: "https://api.openai.com/v1"
    keys:
      - id: "oa-1"
        secret: "${OPENAI_KEY_1}"
        limit_req_per_min: 200
        limit_tokens_per_min: 100000
        pricing:
          input_token_cost: 0.002
          output_token_cost: 0.01
          currency: "USD"
  - id: "gemini"
    name: "Gemini"
    base_url: "https://generativelanguage.googleapis.com/v1"
    keys:
      - id: "gm-1"
        secret: "${GEMINI_KEY_1}"
        limit_req_per_min: 150
        limit_tokens_per_min: 80000
        pricing:
          input_token_cost: 0.00025
          output_token_cost: 0.0005
          currency: "USD"
  - id: "claude"
    name: "Claude"
    base_url: "https://api.anthropic.com/v1"
    keys:
      - id: "cl-1"
        secret: "${CLAUDE_KEY_1}"
        limit_req_per_min: 100
        limit_tokens_per_min: 60000
        pricing:
          input_token_cost: 0.015
          output_token_cost: 0.075
          currency: "USD"

model_aliases:
  gpt-4o: openai:gpt-4o
  gemini-1.5-pro: gemini:gemini-1.5-pro
  claude-3-opus: claude:claude-3-opus

policy:
  strategy: "hybrid"
  cost_first: false
  hybrid_weights:
    token_ratio: 0.3
    req_ratio: 0.2
    error_score: 0.2
    latency: 0.1
    cost_ratio: 0.2
```

## Configuration Sections

### Server Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `listen` | string | `:8080` | Server listen address |
| `admin_api_key` | string | - | API key for admin endpoints |

### Logging Configuration

#### File Logging
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Enable file logging |
| `path` | string | `./logs/llm.log` | Log file path |
| `max_size_mb` | int | `100` | Max file size in MB |
| `max_backups` | int | `5` | Max backup files |

#### Prometheus Logging
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | bool | `true` | Enable Prometheus metrics |
| `endpoint` | string | `/metrics` | Metrics endpoint path |

#### Log Providers
Array of log provider configurations:

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Provider name |
| `type` | string | Provider type (`http`, `prometheus`, etc.) |
| `endpoint` | string | HTTP endpoint for webhooks |
| `batch.enabled` | bool | Enable batching |
| `batch.size` | int | Batch size |
| `batch.interval_seconds` | int | Batch interval |

### Storage Configuration

#### Config Storage
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | string | `file` | Storage type (`file`, `http`) |
| `path` | string | `./configs/config.yaml` | File path (for file type) |

#### Runtime Storage
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | string | `redis` | Storage type (`redis`, `file`, `http`) |
| `addr` | string | `localhost:6379` | Redis address or HTTP endpoint |
| `password` | string | - | Redis password |
| `api_key` | string | - | API key for HTTP storage |

### Provider Configuration

Array of provider configurations:

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique provider ID |
| `name` | string | Human-readable name |
| `base_url` | string | Provider API base URL |

#### Key Configuration
Array of API keys per provider:

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique key ID |
| `secret` | string | API key secret |
| `limit_req_per_min` | int | Request rate limit |
| `limit_tokens_per_min` | int | Token rate limit |

#### Pricing Configuration
| Field | Type | Description |
|-------|------|-------------|
| `input_token_cost` | float64 | Cost per 1K input tokens |
| `output_token_cost` | float64 | Cost per 1K output tokens |
| `currency` | string | Currency code (e.g., "USD") |

### Model Aliases

Map of alias names to provider:model combinations:

```yaml
model_aliases:
  gpt-4: openai:gpt-4
  smart-model: gemini:gemini-1.5-pro
```

### Policy Configuration

Load balancing strategy configuration:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `strategy` | string | `round_robin` | Strategy: `round_robin`, `least_error`, `hybrid` |
| `cost_first` | bool | `false` | Prioritize cost over performance |
| `hybrid_weights.*` | float64 | - | Weights for hybrid scoring |

## Environment Variables

Configuration supports environment variable substitution:

```yaml
providers:
  - keys:
      - secret: "${OPENAI_API_KEY}"
```

Set environment variables before running:

```bash
export OPENAI_API_KEY="sk-your-key"
./truckllm -config config.yaml
```

## Configuration Validation

TruckLLM validates configuration on startup:

- Required fields presence
- URL format validation
- Numeric range checks
- Provider and key uniqueness

Invalid configurations will prevent startup with detailed error messages.

## Hot Reload

Configuration can be reloaded without restarting:

```bash
curl -X POST http://localhost:8080/admin/v1/reload \
  -H "Authorization: Bearer your-admin-key"
```

## Example Configurations

### Minimal Configuration

```yaml
version: "1.0"
server:
  listen: ":8080"
providers:
  - id: openai
    base_url: "https://api.openai.com/v1"
    keys:
      - secret: "sk-your-key"
model_aliases:
  gpt-4: openai:gpt-4
```

### Production Configuration

```yaml
version: "1.0"
server:
  listen: ":8080"
  admin_api_key: "${ADMIN_KEY}"

logging:
  file:
    enabled: true
    path: "/var/log/truckllm/llm.log"
  prometheus:
    enabled: true

storage:
  runtime:
    type: redis
    addr: "redis:6379"
    password: "${REDIS_PASSWORD}"

providers:
  - id: openai
    base_url: "https://api.openai.com/v1"
    keys:
      - id: "key1"
        secret: "${OPENAI_KEY_1}"
        limit_req_per_min: 200
        limit_tokens_per_min: 100000
        pricing:
          input_token_cost: 0.002
          output_token_cost: 0.01
          currency: "USD"
      - id: "key2"
        secret: "${OPENAI_KEY_2}"
        limit_req_per_min: 200
        limit_tokens_per_min: 100000
        pricing:
          input_token_cost: 0.002
          output_token_cost: 0.01
          currency: "USD"

model_aliases:
  gpt-4: openai:gpt-4
  gpt-3.5-turbo: openai:gpt-3.5-turbo

policy:
  strategy: "hybrid"
  hybrid_weights:
    token_ratio: 0.3
    req_ratio: 0.2
    error_score: 0.2
    latency: 0.1
    cost_ratio: 0.2
```

## Configuration API

Manage configuration via REST API:

```bash
# Get current config
curl http://localhost:8080/admin/v1/config

# Update config
curl -X POST http://localhost:8080/admin/v1/config \
  -H "Content-Type: application/json" \
  -d @new-config.json

# Validate config
curl -X POST http://localhost:8080/admin/v1/config/validate \
  -H "Content-Type: application/json" \
  -d @config.json
```