---
sidebar_position: 3
tags: [developer-guide, balancer]
---


# Load Balancer

The load balancer is the core intelligence of COO-LLM, responsible for selecting optimal provider and API key combinations based on performance, cost, and availability.

## Load Balancing Algorithms

### Round Robin

**Algorithm:** `round_robin`

Distributes requests evenly across all available keys in a provider, respecting rate limits.

**Use Case:** Simple load distribution when all keys have similar performance and limits.

**Algorithm:**
1. Filter keys that are not at rate limit
2. If no available keys, use all keys (allow bursting)
3. Select key using round-robin from available keys

### Least Loaded

**Algorithm:** `least_loaded`

Selects the key with the lowest total token usage, preferring non-rate-limited keys.

**Use Case:** Distribute load to avoid rate limits, prioritize underutilized keys.

**Algorithm:**
1. First pass: Find non-rate-limited keys with lowest token usage
2. Second pass: If no non-rate-limited keys, use any key with lowest usage
3. Select key with minimum token usage

### Hybrid

**Algorithm:** `hybrid`

Combines multiple metrics with configurable weights for intelligent key selection.

**Use Case:** Balanced optimization of performance, cost, and reliability.

**Scoring Formula:**
```go
score = w.req_ratio * reqUsage +
        w.token_ratio * tokenUsage +
        w.error_score * errorRate +
        w.latency * avgLatency +
        w.cost_ratio * estCost
```

**Where:**
- `reqUsage`: Total requests processed by key (higher = more used)
- `tokenUsage`: Total tokens processed by key (higher = more used)
- `errorRate`: Error count / total requests (0-1, higher = worse)
- `avgLatency`: Average response time in milliseconds (higher = slower)
- `estCost`: Estimated cost per request in USD (higher = more expensive)

**Lower scores are better** - algorithm selects key with minimum score.

**Priority Presets:**
- `balanced`: All weights = 0.2
- `cost`: cost_ratio = 0.5, others = 0.125
- `req`: req_ratio = 0.5, others = 0.125
- `token`: token_ratio = 0.5, others = 0.125

## Cost Optimization

### Cost Optimization

Use `priority: "cost"` to prioritize cost minimization in hybrid scoring.

**Algorithm:**
1. Set cost_ratio = 0.5 in hybrid weights
2. Calculate cost for each provider/key combination
3. Include cost in overall scoring
4. Select lowest score option

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
- `req`: Requests per key
- `input_tokens`: Input tokens per key
- `output_tokens`: Output tokens per key
- `tokens`: Total tokens per key
- `errors`: Failed requests per key
- `latency`: Average response time per key

**Storage:** Redis/memory with TTL for sliding windows

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
  strategy: "hybrid"          # Legacy field, use algorithm
  algorithm: "hybrid"         # round_robin, least_loaded, hybrid
  priority: "balanced"        # balanced, cost, req, token (auto-sets weights)
  hybrid_weights:             # Manual weights (auto-set based on priority)
    token_ratio: 0.2          # Weight for token usage
    req_ratio: 0.2            # Weight for request usage
    error_score: 0.2          # Weight for error rate
    latency: 0.2              # Weight for response time
    cost_ratio: 0.2           # Weight for cost
  retry:
    max_attempts: 3           # Max retry attempts on failure
    timeout: "30s"            # Timeout per attempt
    interval: "1s"            # Delay between retries
  cache:
    enabled: true             # Enable response caching
    ttl_seconds: 10           # Cache TTL
```

**Priority Options:**
- `balanced`: Equal weights for all metrics (0.2 each)
- `cost`: Prioritize cost minimization (cost_ratio = 0.5, others = 0.125)
- `req`: Prioritize request distribution (req_ratio = 0.5, others = 0.125)
- `token`: Prioritize token efficiency (token_ratio = 0.5, others = 0.125)

### Provider Limits

Rate limits are configured per provider in the `llm_providers` section:

```yaml
llm_providers:
  - id: "openai-prod"
    type: "openai"
    api_keys: ["${OPENAI_KEY_1}"]
    limits:
      req_per_min: 200         # Requests per minute per key
      tokens_per_min: 100000   # Tokens per minute per key
```

## Algorithm Details

### Hybrid Scoring Implementation

The hybrid algorithm calculates scores for each available key using the `calculateScore` method:

```go
func (s *Selector) calculateScore(providerID string, key *config.Key, model string) float64 {
    w := s.cfg.Policy.HybridWeights

    reqUsage, _ := s.store.GetUsage(providerID, key.ID, "req")
    tokenUsage, _ := s.store.GetUsage(providerID, key.ID, "tokens")
    errorScore, _ := s.store.GetUsage(providerID, key.ID, "errors")
    latency, _ := s.store.GetUsage(providerID, key.ID, "latency")

    // Estimate cost (1000 tokens per request)
    avgTokens := 1000.0
    estimatedCost := (key.Pricing.InputTokenCost + key.Pricing.OutputTokenCost) * avgTokens / 1000

    score := w.ReqRatio*reqUsage + w.TokenRatio*tokenUsage + w.ErrorScore*errorScore + w.Latency*latency + w.CostRatio*estimatedCost
    return score
}
```

**Metric Explanations:**
- **req_ratio**: Distributes requests evenly across keys
- **token_ratio**: Prevents keys from hitting token limits
- **error_score**: Avoids unreliable keys (higher error count = higher penalty)
- **latency**: Prefers faster keys (higher latency = higher penalty)
- **cost_ratio**: Minimizes cost when priority is set to "cost"

### Selection Process

1. **Resolve Model:** Map model alias to provider ID using `model_aliases`
2. **Get Provider Config:** Retrieve provider configuration from `llm_providers`
3. **Select Algorithm:** Choose based on `policy.algorithm` setting
4. **Filter Rate Limited Keys:** Skip keys exceeding req/min or tokens/min limits
5. **Calculate Scores:** For hybrid algorithm, compute weighted scores
6. **Select Best Key:** Choose key with lowest score (or random for round-robin)
7. **Fallback:** If no keys available, allow bursting on any key

### Cost Estimation

Cost is estimated using provider pricing and average token usage:

```go
estimatedCost := (inputCost + outputCost) * avgTokensPerRequest / 1000
```

Where:
- `avgTokensPerRequest = 1000` (500 input + 500 output tokens)
- Costs are in USD per 1000 tokens

**Example Calculation:**
- OpenAI GPT-4: $0.03 input + $0.06 output = $0.09 per 1000 tokens
- Gemini Pro: $0.0005 input + $0.0015 output = $0.002 per 1000 tokens
- Hybrid scoring favors cheaper providers when `priority: "cost"`

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

### Algorithm Selection

- **Round Robin:** Simple deployments, uniform key performance
- **Least Loaded:** High-throughput scenarios, avoid hitting limits
- **Hybrid:** Most production use cases, balanced optimization

### Configuration Examples

**Cost-Optimized:**
```yaml
policy:
  algorithm: "hybrid"
  priority: "cost"
```

**Performance-Optimized:**
```yaml
policy:
  algorithm: "hybrid"
  priority: "balanced"
  hybrid_weights:
    latency: 0.3    # Prioritize speed
    cost_ratio: 0.1 # Less cost focus
```

**High-Reliability:**
```yaml
policy:
  algorithm: "hybrid"
  priority: "balanced"
  hybrid_weights:
    error_score: 0.3  # Strong error penalty
```

### Monitoring

- Set up alerts for high error rates (>5%)
- Monitor cost accumulation vs. budget
- Track key utilization distribution (should be ~equal)
- Watch latency percentiles (P95 < 5s)

### Scaling

- Add more keys to increase throughput
- Use multiple providers for redundancy
- Monitor and adjust rate limits based on usage
- Consider `least_loaded` for high-volume deployments

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
- Set `priority: "cost"`
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