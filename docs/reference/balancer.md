# Load Balancer

---
sidebar_position: 3
tags: [developer-guide, balancer]
---

The load balancer is the core intelligence of TruckLLM, responsible for selecting optimal provider and API key combinations based on performance, cost, and availability.

## Load Balancing Strategies

### Round Robin

**Strategy:** `round_robin`

Distributes requests evenly across all available keys in a provider.

**Use Case:** Simple load distribution when all keys have similar performance.

**Algorithm:**
```go
func (s *Selector) selectRoundRobin(pCfg *config.Provider) (*config.Key, error) {
    // Cycle through keys in order
    return &pCfg.Keys[s.roundRobinIndex % len(pCfg.Keys)], nil
}
```

### Least Error

**Strategy:** `least_error`

Prioritizes keys with the lowest error rates.

**Use Case:** Maximizing reliability by avoiding problematic keys.

**Algorithm:**
- Track error count per key over time window
- Select key with lowest error rate
- Fallback to round robin on ties

### Hybrid

**Strategy:** `hybrid`

Combines multiple metrics with configurable weights.

**Use Case:** Balanced optimization of performance, cost, and reliability.

**Scoring Formula:**
```
score = (req_ratio × request_usage) +
        (token_ratio × token_usage) +
        (error_score × error_rate) +
        (latency × avg_latency) +
        (cost_ratio × estimated_cost)
```

Lower scores are better (cost minimization).

## Cost Optimization

### Cost-First Mode

When `cost_first: true`, the balancer prioritizes the cheapest available option.

**Algorithm:**
1. Calculate cost for each provider/key combination
2. Filter combinations within rate limits
3. Select lowest cost option
4. Fall back to performance metrics on ties

### Cost Estimation

Cost is estimated based on:
- Input token cost × estimated input tokens
- Output token cost × estimated output tokens
- Historical usage patterns

### Real-Time Pricing

Pricing data is configured per key:
```yaml
pricing:
  input_token_cost: 0.002    # $ per 1K tokens
  output_token_cost: 0.01    # $ per 1K tokens
  currency: "USD"
```

## Rate Limit Management

### Key Rotation

Automatic key rotation when approaching limits:

1. **Monitor Usage:** Track requests/minute and tokens/minute per key
2. **Threshold Detection:** Alert when usage > 80% of limit
3. **Rotation:** Switch to alternative key
4. **Cooldown:** Allow time for limits to reset

### Provider Failover

When a provider is unavailable:
1. Mark provider as degraded
2. Route traffic to alternative providers
3. Gradually increase traffic as provider recovers

## Metrics Collection

### Usage Tracking

**Runtime Metrics:**
- `req_count`: Requests per key per minute
- `token_count`: Tokens processed per key per minute
- `error_count`: Failed requests per key
- `latency_avg`: Average response time per key

**Storage:** Redis with TTL for sliding windows

### Performance Monitoring

**Collected Metrics:**
- Request latency percentiles
- Error rates by provider/key
- Token throughput
- Cost accumulation

## Configuration

### Policy Configuration

```yaml
policy:
  strategy: "hybrid"          # round_robin, least_error, hybrid
  cost_first: false           # Prioritize cost over performance
  hybrid_weights:
    req_ratio: 0.2           # Weight for request usage
    token_ratio: 0.3         # Weight for token usage
    error_score: 0.2         # Weight for error rate
    latency: 0.1             # Weight for response time
    cost_ratio: 0.2          # Weight for cost
```

### Key Limits

```yaml
keys:
  - id: "key1"
    limit_req_per_min: 200
    limit_tokens_per_min: 100000
```

## Algorithm Details

### Hybrid Scoring

For each available key, calculate a weighted score:

```go
func calculateScore(providerID, keyID string) float64 {
    reqUsage := getUsage(providerID, keyID, "req")
    tokenUsage := getUsage(providerID, keyID, "tokens")
    errorRate := getUsage(providerID, keyID, "errors")
    avgLatency := getUsage(providerID, keyID, "latency")
    estCost := estimateCost(keyID)

    score := w.req_ratio * reqUsage +
            w.token_ratio * tokenUsage +
            w.error_score * errorRate +
            w.latency * avgLatency +
            w.cost_ratio * estCost

    return score
}
```

### Selection Process

1. **Filter Available Keys:** Remove keys exceeding rate limits
2. **Calculate Scores:** Compute scores for remaining keys
3. **Select Optimal:** Choose key with best (lowest) score
4. **Fallback:** Round robin if no clear winner

### Cost Estimation

```go
func estimateCost(key *config.Key) float64 {
    avgTokens := 1000.0  // Estimated tokens per request
    inputCost := key.Pricing.InputTokenCost * avgTokens / 1000
    outputCost := key.Pricing.OutputTokenCost * avgTokens / 1000
    return inputCost + outputCost
}
```

## Monitoring & Observability

### Admin API

View balancer status:

```bash
curl http://localhost:8080/admin/v1/providers
```

Response includes:
- Key usage statistics
- Error rates
- Performance metrics
- Current selection status

### Prometheus Metrics

```prometheus
# Request distribution
llm_requests_total{provider="openai",key="key1"} 150

# Performance metrics
llm_request_duration_seconds{provider="openai",quantile="0.95"} 2.5

# Cost tracking
llm_cost_total{provider="openai",currency="USD"} 12.50
```

## Best Practices

### Strategy Selection

- **Round Robin:** Simple deployments, uniform key performance
- **Least Error:** High reliability requirements
- **Hybrid:** Balanced performance and cost optimization

### Weight Tuning

Start with equal weights and adjust based on:
- Business priorities (cost vs. performance)
- Observed metrics
- SLA requirements

### Monitoring

- Set up alerts for high error rates
- Monitor cost accumulation
- Track key utilization distribution

### Scaling

- Add more keys to increase throughput
- Use multiple providers for redundancy
- Monitor and adjust rate limits

## Troubleshooting

### Common Issues

**All keys at limit:**
- Increase rate limits in provider dashboard
- Add more API keys
- Implement request queuing

**High latency:**
- Check provider status
- Switch to closer regions
- Optimize request batching

**Cost spikes:**
- Review pricing configuration
- Enable cost-first mode
- Set spending limits

### Debug Commands

```bash
# View current key usage
curl http://localhost:8080/admin/v1/providers

# Check configuration
curl http://localhost:8080/admin/v1/config

# View recent logs
curl "http://localhost:8080/admin/v1/logs?limit=50"
```

## Future Enhancements

- **Machine Learning:** Predictive key selection based on historical patterns
- **Dynamic Limits:** Adjust limits based on provider announcements
- **Multi-Region:** Geographic load balancing
- **Request Batching:** Combine multiple requests for efficiency