---
sidebar_position: 2
tags: [reference, errors]
---

# Error Codes & Troubleshooting

Complete reference of error codes and troubleshooting steps.

## HTTP Status Codes

### 2xx Success

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created |

### 4xx Client Errors

| Code | Error | Description | Troubleshooting |
|------|-------|-------------|----------------|
| 400 | Bad Request | Invalid request format | Check JSON syntax, required fields |
| 401 | Unauthorized | Invalid API key | Verify `Authorization: Bearer <key>` header |
| 403 | Forbidden | Provider not allowed | Check `allowed_providers` in config |
| 404 | Not Found | Endpoint/model not found | Verify URL and model name |
| 429 | Too Many Requests | Rate limit exceeded | Wait and retry with backoff |

### 5xx Server Errors

| Code | Error | Description | Troubleshooting |
|------|-------|-------------|----------------|
| 500 | Internal Server Error | Server error | Check server logs |
| 502 | Bad Gateway | Provider API error | Check provider status |
| 503 | Service Unavailable | Server overloaded | Wait and retry |
| 504 | Gateway Timeout | Request timeout | Increase timeout, check network |

## Application Errors

### Authentication & Authorization

| Error Message | HTTP Code | Cause | Solution |
|---------------|-----------|-------|----------|
| `invalid api key` | 401 | API key not in config | Add key to `api_keys` section |
| `api key not found` | 401 | Key doesn't exist | Check key spelling |
| `provider not allowed` | 403 | Key restricted from provider | Update `allowed_providers` |
| `model not found` | 404 | Invalid model name | Use `provider:model` format |

### Rate Limiting

| Error Message | HTTP Code | Cause | Solution |
|---------------|-----------|-------|----------|
| `rate limit exceeded` | 429 | Too many requests | Wait, reduce request rate |
| `token limit exceeded` | 429 | Too many tokens used | Check usage, add more keys |
| `session limit exceeded` | 429 | Session quota reached | Wait for session reset |

### Provider Errors

| Error Message | HTTP Code | Cause | Solution |
|---------------|-----------|-------|----------|
| `provider error` | 500 | Provider API failed | Check provider status page |
| `invalid provider response` | 500 | Unexpected API response | Update provider integration |
| `provider timeout` | 504 | Provider slow/unavailable | Switch providers, increase timeout |

### Configuration Errors

| Error Message | HTTP Code | Cause | Solution |
|---------------|-----------|-------|----------|
| `config validation failed` | 500 | Invalid config on startup | Fix config, restart |
| `missing required field` | 500 | Required config missing | Add missing fields |
| `invalid config format` | 500 | YAML syntax error | Validate YAML |

## Common Issues & Solutions

### Connection Issues

**Problem**: `connection refused`
```
Failed to connect to provider API
```

**Solutions**:
- Check network connectivity
- Verify provider API endpoints
- Check firewall/proxy settings
- Test with `curl https://api.openai.com`

### Configuration Issues

**Problem**: Server won't start
```
config validation failed: missing admin_api_key
```

**Solutions**:
- Add required fields to config
- Check YAML syntax
- Validate environment variables
- Use `coo-llm --config config.yaml --validate`

### Performance Issues

**Problem**: High latency
```
Request taking >10 seconds
```

**Solutions**:
- Check provider API status
- Monitor server resources (CPU, memory)
- Switch to different provider
- Enable caching in config

### Memory Issues

**Problem**: Out of memory
```
runtime: out of memory: cannot allocate ...
```

**Solutions**:
- Switch storage from `sql` to `redis`
- Reduce cache TTL
- Monitor goroutine leaks
- Increase container memory limits

## Error Response Format

All errors follow this JSON format:

```json
{
  "error": {
    "message": "rate limit exceeded",
    "type": "rate_limit_error",
    "code": 429,
    "details": {
      "retry_after": 60,
      "limit": 200,
      "remaining": 0
    }
  }
}
```

## Logging Error Details

Enable debug logging for detailed error information:

```yaml
logging:
  level: "debug"
```

Logs include:
- Request ID for tracing
- Stack traces for panics
- Provider response details
- Rate limit counters

## Monitoring Errors

### Prometheus Metrics

```promql
# Error rate by provider
rate(coo_llm_errors_total[5m]) / rate(coo_llm_requests_total[5m])

# Errors by type
coo_llm_errors_total{type="rate_limit"}

# Error rate over time
increase(coo_llm_errors_total[1h])
```

### Alerting

Recommended alerts:
```yaml
- alert: HighErrorRate
  expr: rate(coo_llm_errors_total[5m]) / rate(coo_llm_requests_total[5m]) > 0.1
  for: 5m

- alert: ProviderDown
  expr: up{job="coo-llm"} == 0
  for: 1m
```

## Getting Support

For persistent errors:

1. **Collect diagnostics**:
   - Full error logs
   - Config file (redact secrets)
   - System info: `go version`, `uname -a`

2. **Test isolation**:
   - Try with minimal config
   - Test provider APIs directly
   - Check network connectivity

3. **Report issues**:
   - GitHub Issues with reproduction steps
   - Include error logs and config samples