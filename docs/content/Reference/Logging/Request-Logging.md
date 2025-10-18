---
sidebar_position: 1
tags: [reference, logging, requests]
---

# Request Logging

COO-LLM provides comprehensive request logging with structured JSON output for monitoring and debugging.

## Log Structure

Request logs use consistent JSON format with detailed information:

```json
{
  "level": "info",
  "entry": {
    "timestamp": "2025-10-18T10:10:37+07:00",
    "provider": "gemini-prod",
    "model": "gemini-prod:gemini-2.0-flash",
    "req_id": "1760757037738488000",
    "latency_ms": 460,
    "status": 500,
    "tokens": 0,
    "cost": 0,
    "error": "Gemini API error after 3 attempts: googleapi: Error 400: API key not valid..."
  },
  "time": "2025-10-18T10:10:37+07:00",
  "message": "request"
}
```

## Field Descriptions

### Core Fields
- **level**: Log level (`info`, `debug`, `error`)
- **time**: ISO 8601 timestamp when log was written
- **message**: Log message type (`request`)

### Entry Fields
- **timestamp**: Request start time
- **provider**: LLM provider ID (e.g., `openai`, `gemini-prod`)
- **model**: Full model name including provider prefix
- **req_id**: Unique request identifier (Unix nanoseconds)
- **latency_ms**: Total request duration in milliseconds
- **status**: HTTP response status code
- **tokens**: Token usage (input + output tokens)
- **cost**: Estimated cost for the request
- **error**: Error message (if request failed)

## Log Levels

### INFO Level
- Successful requests
- Normal API operations
- Usage tracking

### ERROR Level
- Failed requests
- Provider errors
- Authentication failures
- Rate limiting

### DEBUG Level
- Detailed operation logs
- Store operations
- Internal metrics

## Configuration

```yaml
logging:
  file:
    enabled: true
    path: "./logs/llm.log"
    max_size_mb: 100
    max_backups: 5
  level: "info"  # "debug", "info", "warn", "error"
```

## Log Analysis

### Common Patterns

**Successful Request:**
```json
{
  "level": "info",
  "entry": {
    "provider": "openai",
    "model": "openai:gpt-4o",
    "latency_ms": 1250,
    "status": 200,
    "tokens": 150,
    "cost": 0.00225
  }
}
```

**Failed Request:**
```json
{
  "level": "info", 
  "entry": {
    "provider": "gemini-prod",
    "status": 500,
    "error": "API key not valid"
  }
}
```

### Monitoring Queries

```bash
# Count errors by provider
grep '"level":"info"' logs/llm.log | grep '"status":500' | jq -r '.entry.provider' | sort | uniq -c

# Average latency
grep '"level":"info"' logs/llm.log | jq '.entry.latency_ms' | awk '{sum+=$1; count++} END {print sum/count}'

# Total cost per day
grep '"level":"info"' logs/llm.log | jq '.entry.cost' | paste -sd+ | bc
```
