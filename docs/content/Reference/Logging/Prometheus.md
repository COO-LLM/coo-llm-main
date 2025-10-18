---
sidebar_position: 3
tags: [reference, logging, prometheus, metrics]
---

# Prometheus Metrics

COO-LLM exposes comprehensive metrics for monitoring via Prometheus-compatible endpoint.

## Endpoint

Metrics are available at `/api/metrics` when enabled.

## Configuration

```yaml
logging:
  prometheus:
    enabled: true
    endpoint: "/api/metrics"
```

## Available Metrics

### Request Metrics

- **llm_requests_total**: Total number of LLM requests
  - Type: Counter
  - Labels: `provider`, `model`, `key`, `client_key`
  - Example: `llm_requests_total{provider="openai",model="gpt-4o"} 1250`

- **llm_request_duration_seconds**: Request duration in seconds
  - Type: Histogram
  - Labels: `provider`, `model`, `key`, `client_key`
  - Buckets: 0.1, 0.5, 1, 2.5, 5, 10, 30, 60
  - Example: `llm_request_duration_seconds_bucket{provider="openai",le="1"} 1200`

### Token Metrics

- **llm_tokens_total**: Total token usage
  - Type: Counter
  - Labels: `provider`, `model`, `key`, `client_key`, `type` (input/output)
  - Example: `llm_tokens_total{provider="openai",type="input"} 150000`

### Cost Metrics

- **llm_cost_total**: Total cost incurred
  - Type: Counter
  - Labels: `provider`, `model`, `key`, `client_key`
  - Example: `llm_cost_total{provider="openai"} 25.50`

### Error Metrics

- **llm_errors_total**: Total number of errors
  - Type: Counter
  - Labels: `provider`, `model`, `key`, `client_key`, `type`
  - Example: `llm_errors_total{provider="openai",type="rate_limit"} 5`

### Active Connections

- **llm_active_connections**: Current active connections
  - Type: Gauge
  - Labels: `provider`
  - Example: `llm_active_connections{provider="openai"} 3`

## Query Examples

### Request Rate
```promql
# Requests per second by provider
rate(llm_requests_total[5m])

# Requests per minute
rate(llm_requests_total[1m])
```

### Error Rate
```promql
# Error percentage
rate(llm_errors_total[5m]) / rate(llm_requests_total[5m]) * 100
```

### Latency
```promql
# 95th percentile latency
histogram_quantile(0.95, rate(llm_request_duration_seconds_bucket[5m]))

# Average latency
rate(llm_request_duration_seconds_sum[5m]) / rate(llm_request_duration_seconds_count[5m])
```

### Cost Tracking
```promql
# Cost per hour
increase(llm_cost_total[1h])

# Cost by provider
sum(increase(llm_cost_total[1h])) by (provider)
```

### Token Usage
```promql
# Tokens per minute
rate(llm_tokens_total[1m])

# Input vs output tokens
rate(llm_tokens_total{type="input"}[5m]) / rate(llm_tokens_total{type="output"}[5m])
```

## Alerting Rules

```yaml
groups:
  - name: coo-llm
    rules:
      - alert: HighErrorRate
        expr: rate(llm_errors_total[5m]) / rate(llm_requests_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"

      - alert: HighLatency
        expr: histogram_quantile(0.95, rate(llm_request_duration_seconds_bucket[5m])) > 30
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency detected"

      - alert: RateLimitHit
        expr: increase(llm_errors_total{type="rate_limit"}[5m]) > 10
        for: 2m
        labels:
          severity: info
        annotations:
          summary: "Rate limit exceeded"
```

## Dashboard Examples

### Grafana Dashboard JSON

```json
{
  "dashboard": {
    "title": "COO-LLM Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(llm_requests_total[5m])",
            "legendFormat": "{{provider}}"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph", 
        "targets": [
          {
            "expr": "rate(llm_errors_total[5m]) / rate(llm_requests_total[5m]) * 100",
            "legendFormat": "{{provider}} error %"
          }
        ]
      },
      {
        "title": "Latency P95",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(llm_request_duration_seconds_bucket[5m]))",
            "legendFormat": "{{provider}} p95"
          }
        ]
      }
    ]
  }
}
```

## Integration

### Prometheus Configuration

```yaml
scrape_configs:
  - job_name: 'coo-llm'
    static_configs:
      - targets: ['localhost:2906']
    metrics_path: '/api/metrics'
```

### Grafana Data Source

```json
{
  "name": "COO-LLM Prometheus",
  "type": "prometheus",
  "url": "http://prometheus:9090",
  "access": "proxy"
}
```
