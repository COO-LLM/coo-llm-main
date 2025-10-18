---
sidebar_position: 2
tags: [reference, logging, usage, metrics]
---

# Usage Logging

Detailed logging of internal operations including store interactions and usage tracking.

## Usage Log Structure

```json
{
  "level": "debug",
  "operation": "IncrementUsage",
  "provider": "gemini-prod",
  "keyID": "gemini-prod-9987861163793b2d",
  "metric": "req",
  "delta": 1,
  "time": "2025-10-18T10:10:37+07:00",
  "message": "store operation"
}
```

## Field Descriptions

- **level**: Always `debug` for usage logs
- **operation**: Store operation type
  - `IncrementUsage`: Usage counter increment
  - `SetUsage`: Usage value setting
  - `GetUsage`: Usage value retrieval
- **provider**: Provider identifier
- **keyID**: Hashed key identifier
- **metric**: Usage metric type
  - `req`: API requests
  - `tokens`: Token usage
  - `errors`: Error count
  - `latency`: Response time
  - `cost`: Cost tracking
- **delta**: Change amount (for increments)
- **time**: Operation timestamp
- **message**: Always `"store operation"`

## Metrics Types

### Request Metrics
```json
{
  "operation": "IncrementUsage",
  "metric": "req",
  "delta": 1
}
```

### Token Metrics
```json
{
  "operation": "IncrementUsage", 
  "metric": "tokens",
  "delta": 150
}
```

### Error Metrics
```json
{
  "operation": "IncrementUsage",
  "metric": "errors", 
  "delta": 1
}
```

### Latency Metrics
```json
{
  "operation": "IncrementUsage",
  "metric": "latency",
  "delta": 1250
}
```

## Usage Analysis

### Aggregation Queries

```bash
# Total requests by provider
grep '"operation":"IncrementUsage"' logs/llm.log | grep '"metric":"req"' | jq -r '.provider' | sort | uniq -c

# Total tokens used
grep '"operation":"IncrementUsage"' logs/llm.log | grep '"metric":"tokens"' | jq '.delta' | paste -sd+ | bc

# Error rate calculation
total_req=$(grep '"metric":"req"' logs/llm.log | wc -l)
total_err=$(grep '"metric":"errors"' logs/llm.log | wc -l)
echo "scale=2; $total_err / $total_req * 100" | bc
```

### Real-time Monitoring

```bash
# Watch live usage
tail -f logs/llm.log | grep '"operation":"IncrementUsage"' | jq '.'

# Monitor specific provider
tail -f logs/llm.log | grep '"provider":"openai"' | grep '"metric":"req"'
```

## Configuration

Usage logs are always enabled at debug level:

```yaml
logging:
  level: "debug"  # Required for usage logs
```

## Performance Impact

- **Low Overhead**: Efficient JSON serialization
- **Async Writing**: Non-blocking log operations
- **Buffered Output**: Batched writes for performance
- **Configurable**: Can be disabled by log level
