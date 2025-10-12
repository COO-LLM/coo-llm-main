# Logging

---
sidebar_position: 5
tags: [developer-guide, logging]
---

TruckLLM provides comprehensive logging and observability features with multiple output backends and structured logging capabilities.

## Logging Architecture

```
Application Events
        ↓
    Log Entry Creation
        ↓
    Structured Formatting
        ↓
   ┌─────────────────────┐
   │   Log Processors    │
   │ ┌─────────────────┐ │
   │ │ File Writer     │ │
   │ ├─────────────────┤ │
   │ │ HTTP Webhook    │ │
   │ ├─────────────────┤ │
   │ │ Prometheus      │ │
   │ └─────────────────┘ │
   └─────────────────────┘
```

## Log Levels

- **DEBUG:** Detailed debugging information
- **INFO:** General operational messages
- **WARN:** Warning conditions
- **ERROR:** Error conditions
- **FATAL:** Critical errors causing shutdown

## Structured Logging

All logs use structured JSON format:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "info",
  "provider": "openai",
  "model": "gpt-4",
  "key_id": "oa-1",
  "request_id": "req-123456",
  "latency_ms": 1200,
  "status_code": 200,
  "tokens_used": 150,
  "cost_usd": 0.003,
  "user_agent": "OpenAI/1.0",
  "ip": "192.168.1.100",
  "message": "Request completed successfully"
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
    compress: true
```

### Prometheus Logging

```yaml
logging:
  prometheus:
    enabled: true
    endpoint: "/metrics"
    pushgateway:
      enabled: false
      url: "http://pushgateway:9091"
```

### HTTP Webhook Logging

```yaml
logging:
  providers:
    - name: "logstash"
      type: "http"
      endpoint: "https://logstash.example.com"
      headers:
        Authorization: "Bearer token"
      batch:
        enabled: true
        size: 50
        interval_seconds: 10
```

## Log Types

### Request Logs

Captured for every API request:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "info",
  "event": "request",
  "request_id": "req-123456",
  "method": "POST",
  "path": "/v1/chat/completions",
  "provider": "openai",
  "model": "gpt-4",
  "key_id": "oa-1",
  "input_tokens": 25,
  "output_tokens": 150,
  "total_tokens": 175,
  "latency_ms": 1200,
  "status_code": 200,
  "cost_usd": 0.0035,
  "user_id": "user-123",
  "client_ip": "192.168.1.100"
}
```

### Error Logs

Detailed error information:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "error",
  "event": "error",
  "request_id": "req-123456",
  "error_type": "provider_error",
  "error_code": 429,
  "error_message": "Rate limit exceeded",
  "provider": "openai",
  "key_id": "oa-1",
  "retry_count": 2,
  "stack_trace": "..."
}
```

### System Logs

Application lifecycle events:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "info",
  "event": "startup",
  "version": "1.0.0",
  "config_path": "./configs/config.yaml",
  "providers_loaded": 3,
  "keys_loaded": 5
}
```

## Log Processing

### Batching

For high-throughput scenarios:

```yaml
batch:
  enabled: true
  size: 100          # Max entries per batch
  interval_seconds: 30  # Max time between flushes
  compression: "gzip"
```

### Filtering

Log filtering by level, provider, or custom rules:

```yaml
filters:
  - level: "debug"
    providers: ["openai"]
  - exclude:
      paths: ["/health", "/metrics"]
```

### Sampling

Reduce log volume for high-frequency events:

```yaml
sampling:
  requests:
    rate: 0.1  # Log 10% of requests
  errors:
    rate: 1.0  # Log all errors
```

## Output Backends

### File Backend

**Features:**
- Automatic rotation
- Compression
- Size and time-based rotation

**Configuration:**
```yaml
file:
  enabled: true
  path: "/var/log/truckllm/app.log"
  max_size_mb: 100
  max_backups: 10
  compress: true
  rotation_time: "daily"  # hourly, daily, weekly
```

### HTTP Webhook Backend

**Features:**
- RESTful API integration
- Custom headers
- Retry logic
- Batching support

**Configuration:**
```yaml
providers:
  - name: "elasticsearch"
    type: "http"
    endpoint: "https://es.example.com/_bulk"
    method: "POST"
    headers:
      Content-Type: "application/x-ndjson"
      Authorization: "ApiKey key"
    timeout_seconds: 30
    retry:
      max_attempts: 3
      backoff_multiplier: 2.0
```

### Prometheus Backend

**Features:**
- Metrics export
- Histogram buckets
- Custom labels

**Exported Metrics:**
```prometheus
# Request metrics
llm_requests_total{provider="openai",model="gpt-4",status="200"} 1250
llm_request_duration_seconds{provider="openai",quantile="0.95"} 2.5

# Cost metrics
llm_cost_total{provider="openai",currency="USD"} 12.50

# Error metrics
llm_errors_total{provider="openai",error_type="rate_limit"} 5
```

## Log Analysis

### Log Queries

**Find slow requests:**
```bash
grep '"latency_ms":[0-9]\{4,\}' logs/llm.log
```

**Count requests by provider:**
```bash
grep '"event":"request"' logs/llm.log | jq -r '.provider' | sort | uniq -c
```

**Find errors:**
```bash
grep '"level":"error"' logs/llm.log
```

### Monitoring Dashboards

**Grafana Dashboard:**
- Request rate by provider
- Error rate trends
- Latency percentiles
- Cost accumulation

**Alerting Rules:**
- High error rate (>5%)
- Increased latency (>2s p95)
- Rate limit hits
- Cost budget exceeded

## Performance

### Throughput

- **File logging:** 10,000+ logs/second
- **HTTP logging:** 1,000+ logs/second (with batching)
- **Memory usage:** <50MB for 1M buffered logs

### Asynchronous Processing

All logging is asynchronous to avoid blocking request processing:

```go
// Non-blocking log submission
go func() {
    logger.LogRequest(ctx, entry)
}()
```

### Buffer Management

- Circular buffers for high throughput
- Automatic flush on buffer full
- Graceful shutdown with flush

## Security

### Sensitive Data

- API keys never logged
- PII data masked
- Request bodies truncated for large payloads

### Audit Logging

- Authentication events
- Configuration changes
- Administrative actions

### Compliance

- GDPR compliance for EU users
- SOC 2 logging requirements
- Data retention policies

## Best Practices

### Log Levels

- **DEBUG:** Development only
- **INFO:** Normal operations
- **WARN:** Degraded performance
- **ERROR:** Failures requiring attention
- **FATAL:** System shutdown

### Structured Fields

Always include:
- `timestamp`
- `request_id` (correlation)
- `level`
- `event` type
- `provider` and `key_id`

### Retention

- **Application logs:** 30 days
- **Audit logs:** 1 year
- **Metrics:** 90 days
- **Archives:** Compressed long-term storage

## Troubleshooting

### Common Issues

**Logs not appearing:**
- Check file permissions
- Verify configuration
- Check disk space

**High latency:**
- Disable synchronous logging
- Use batching
- Check network for HTTP backends

**Log loss:**
- Implement retry logic
- Use durable queues
- Monitor buffer usage

### Debug Mode

Enable debug logging:

```bash
export LOG_LEVEL=debug
./truckllm -config config.yaml
```

### Log Inspection

```bash
# Tail logs
tail -f logs/llm.log | jq .

# Search for specific request
grep "req-123456" logs/llm.log

# Count errors by hour
jq -r '.timestamp[:13]' logs/llm.log | sort | uniq -c
```

## Integration Examples

### ELK Stack

```yaml
providers:
  - name: "logstash"
    type: "http"
    endpoint: "https://logstash:8080"
    batch:
      enabled: true
      size: 100
```

### Splunk

```yaml
providers:
  - name: "splunk"
    type: "http"
    endpoint: "https://splunk.example.com:8088/services/collector"
    headers:
      Authorization: "Splunk token"
```

### CloudWatch

```yaml
providers:
  - name: "cloudwatch"
    type: "http"
    endpoint: "https://logs.amazonaws.com/"
    # Custom implementation needed
```

## Future Enhancements

- **Log Shipping:** Direct integration with log aggregation services
- **Distributed Tracing:** OpenTelemetry integration
- **Log Analytics:** Built-in query and visualization
- **Anomaly Detection:** ML-based log analysis
- **Compliance Automation:** Automated log retention and deletion