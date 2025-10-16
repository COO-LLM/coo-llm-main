---
sidebar_position: 5
tags: [developer-guide, logging]
---

# Logging

COO-LLM provides structured logging with file output and Prometheus metrics integration.

## Log Implementation

Based on `internal/log/logger.go`, COO-LLM uses Zerolog for structured JSON logging with the following features:

- File logging with rotation
- Prometheus metrics integration
- Request logging with context
- Structured JSON format

## Log Levels

- **INFO:** Normal operations and request logs
- **WARN:** Warnings and non-critical issues
- **ERROR:** Errors and failures
- **DEBUG:** Detailed debugging (when enabled)

## Structured Logging

Logs use JSON format with consistent fields:

```json
{
  "level": "info",
  "provider": "openai-prod",
  "model": "gpt-4o",
  "reqID": "1234567890",
  "latencyMS": 1200,
  "status": 200,
  "tokens": 150,
  "error": "",
  "time": "2024-01-01T12:00:00Z"
}
```

## Configuration

### File Logging

```yaml
logging:
  file:
    enabled: true
    path: "./logs/llm.log"
    max_size_mb: 100
    max_backups: 5
```

### Prometheus Metrics

```yaml
logging:
  prometheus:
    enabled: true
    endpoint: "/metrics"
```

**Note:** HTTP webhooks and advanced batching are not yet implemented.

## Log Types

### Request Logs

Logged for each chat completion request via `logger.LogRequest()`:

```json
{
  "level": "info",
  "provider": "openai-prod",
  "model": "gpt-4o",
  "reqID": "1699123456789",
  "latencyMS": 1200,
  "status": 200,
  "tokens": 150,
  "error": "",
  "time": "2024-01-01T12:00:00Z"
}
```

**Fields:**
- `provider`: Provider ID used for the request
- `model`: Model name requested
- `reqID`: Unique request identifier
- `latencyMS`: Response time in milliseconds
- `status`: HTTP status code
- `tokens`: Total tokens used
- `error`: Error message (if any)

### Application Logs

General application events and errors:

```json
{
  "level": "info",
  "message": "Starting server on :2906",
  "time": "2024-01-01T12:00:00Z"
}
```

### Storage Operation Logs

All storage backends log database operations for monitoring and debugging:

```json
{
  "level": "debug",
  "operation": "GetUsage",
  "provider": "openai",
  "keyID": "key1",
  "metric": "req",
  "value": 45,
  "message": "store operation",
  "time": "2024-01-01T12:00:00Z"
}
```

```json
{
  "level": "debug",
  "operation": "IncrementUsage",
  "provider": "openai",
  "keyID": "key1",
  "metric": "tokens",
  "delta": 150,
  "old_value": 1200,
  "new_value": 1350,
  "message": "store operation",
  "time": "2024-01-01T12:00:00Z"
}
```

**Storage Log Fields:**
- `operation`: Operation type (GetUsage, SetUsage, IncrementUsage, etc.)
- `provider`: Provider identifier
- `keyID`: API key identifier
- `metric`: Metric name (req, tokens, etc.)
- `value`/`delta`: Numeric values involved
- `error`: Error message for failed operations

### Prometheus Metrics

When enabled, metrics are exposed at `/metrics` endpoint.

## Output Backends

### File Backend

**Features:**
- JSON structured logging
- Automatic rotation by size
- Configurable retention

**Configuration:**
```yaml
logging:
  file:
    enabled: true
    path: "./logs/llm.log"
    max_size_mb: 100
    max_backups: 5
```

### Prometheus Backend

**Features:**
- Metrics endpoint at `/metrics`
- Integration with monitoring systems

**Configuration:**
```yaml
logging:
  prometheus:
    enabled: true
    endpoint: "/metrics"
```

**Note:** HTTP webhooks, batching, filtering, and advanced features are not yet implemented.

## Log Analysis

### Log Queries

**Find slow requests:**
```bash
grep '"latencyMS":[0-9]\{4,\}' logs/llm.log
```

**Count requests by provider:**
```bash
grep '"provider"' logs/llm.log | jq -r '.provider' | sort | uniq -c
```

**Find errors:**
```bash
grep '"level":"error"' logs/llm.log
```

### Monitoring

**Prometheus metrics can be scraped for monitoring:**
```prometheus
# Metrics are available at /metrics when enabled
```

## Best Practices

### Log Levels

- **INFO:** Normal request logging
- **WARN:** Warnings
- **ERROR:** Errors and failures

### Structured Fields

Request logs include:
- `provider`: Provider used
- `model`: Model requested
- `reqID`: Request correlation ID
- `latencyMS`: Response time
- `status`: HTTP status
- `tokens`: Token usage
- `error`: Error details

### Retention

- **Application logs:** Configurable via `max_backups`
- **Log rotation:** Automatic by size (`max_size_mb`)

## Troubleshooting

### Common Issues

**Logs not appearing:**
- Check file permissions on `./logs/` directory
- Verify `logging.file.enabled: true`
- Check available disk space

**Log rotation not working:**
- Ensure write permissions on log directory
- Check `max_size_mb` and `max_backups` settings

### Log Inspection

```bash
# View recent logs
tail -f logs/llm.log | jq .

# Search for specific request ID
grep "1234567890" logs/llm.log

# Count requests by status
jq -r '.status' logs/llm.log | sort | uniq -c
```

## Implementation Notes

- Logging is synchronous to ensure request logs are written
- Uses Zerolog for high-performance structured logging
- File rotation prevents disk space issues
- Prometheus integration provides metrics endpoint