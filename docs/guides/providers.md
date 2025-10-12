# Providers

---
sidebar_position: 4
tags: [user-guide, providers]
---

TruckLLM supports multiple LLM providers through a plugin-based architecture. Each provider implements a common interface for seamless integration.

## Supported Providers

### OpenAI

**Provider ID:** `openai`

**Configuration:**
```yaml
providers:
  - id: openai
    name: "OpenAI"
    base_url: "https://api.openai.com/v1"
    keys:
      - id: "oa-1"
        secret: "${OPENAI_API_KEY}"
        limit_req_per_min: 200
        limit_tokens_per_min: 100000
        pricing:
          input_token_cost: 0.002
          output_token_cost: 0.01
          currency: "USD"
```

**Supported Models:**
- `gpt-4`
- `gpt-4-turbo`
- `gpt-4o`
- `gpt-3.5-turbo`
- `gpt-3.5-turbo-instruct`
- `text-embedding-ada-002`

**Rate Limits:** Based on OpenAI tier (see [OpenAI docs](https://platform.openai.com/docs/guides/rate-limits))

### Google Gemini

**Provider ID:** `gemini`

**Configuration:**
```yaml
providers:
  - id: gemini
    name: "Gemini"
    base_url: "https://generativelanguage.googleapis.com/v1"
    keys:
      - id: "gm-1"
        secret: "${GEMINI_API_KEY}"
        limit_req_per_min: 60
        limit_tokens_per_min: 32000
        pricing:
          input_token_cost: 0.00025
          output_token_cost: 0.0005
          currency: "USD"
```

**Supported Models:**
- `gemini-1.5-pro`
- `gemini-1.5-flash`
- `gemini-pro`

**Rate Limits:** 60 requests/minute for free tier, higher for paid (see [Gemini docs](https://ai.google.dev/docs/rate-limits))

### Anthropic Claude

**Provider ID:** `claude`

**Configuration:**
```yaml
providers:
  - id: claude
    name: "Claude"
    base_url: "https://api.anthropic.com/v1"
    keys:
      - id: "cl-1"
        secret: "${CLAUDE_API_KEY}"
        limit_req_per_min: 50
        limit_tokens_per_min: 25000
        pricing:
          input_token_cost: 0.015
          output_token_cost: 0.075
          currency: "USD"
```

**Supported Models:**
- `claude-3-opus-20240229`
- `claude-3-sonnet-20240229`
- `claude-3-haiku-20240307`
- `claude-2.1`

**Rate Limits:** Varies by model (see [Anthropic docs](https://docs.anthropic.com/claude/docs/rate-limits))

## Provider Interface

All providers implement the `Provider` interface:

```go
type Provider interface {
    Name() string
    Generate(ctx context.Context, req *Request) (*Response, error)
    ListModels(ctx context.Context) ([]string, error)
}
```

### Request Structure

```go
type Request struct {
    Model  string                 `json:"model"`
    Input  map[string]interface{} `json:"input"`
    APIKey string                 `json:"api_key"`
}
```

### Response Structure

```go
type Response struct {
    RawResponse []byte
    HTTPCode    int
    Err         error
    TokensUsed  int
    Latency     int64 // milliseconds
}
```

## Adding Custom Providers

To add a new provider:

1. **Implement the Provider interface:**

```go
package provider

import (
    "context"
    "encoding/json"
    "net/http"
)

type CustomProvider struct {
    cfg *config.Provider
}

func NewCustomProvider(cfg *config.Provider) *CustomProvider {
    return &CustomProvider{cfg: cfg}
}

func (p *CustomProvider) Name() string {
    return p.cfg.ID
}

func (p *CustomProvider) Generate(ctx context.Context, req *Request) (*Response, error) {
    // Transform request to provider format
    providerReq := transformRequest(req)

    // Make HTTP request
    resp, err := makeRequest(p.cfg.BaseURL, providerReq, req.APIKey)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Parse response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    // Extract token usage
    tokensUsed := extractTokens(body)

    return &Response{
        RawResponse: body,
        HTTPCode:    resp.StatusCode,
        TokensUsed:  tokensUsed,
        Latency:     calculateLatency(startTime),
    }, nil
}

func (p *CustomProvider) ListModels(ctx context.Context) ([]string, error) {
    // Return supported models
    return []string{"model-1", "model-2"}, nil
}
```

2. **Register in the registry:**

```go
func (r *Registry) LoadFromConfig(cfg *config.Config) error {
    for _, pCfg := range cfg.Providers {
        var p Provider
        switch pCfg.ID {
        case "openai":
            p = NewOpenAIProvider(&pCfg)
        case "gemini":
            p = NewGeminiProvider(&pCfg)
        case "claude":
            p = NewClaudeProvider(&pCfg)
        case "custom":
            p = NewCustomProvider(&pCfg)
        default:
            return fmt.Errorf("unsupported provider: %s", pCfg.ID)
        }
        r.Register(p)
    }
    return nil
}
```

3. **Add configuration:**

```yaml
providers:
  - id: custom
    name: "Custom Provider"
    base_url: "https://api.custom-provider.com/v1"
    keys:
      - id: "custom-1"
        secret: "${CUSTOM_API_KEY}"
        limit_req_per_min: 100
        limit_tokens_per_min: 50000
        pricing:
          input_token_cost: 0.001
          output_token_cost: 0.002
          currency: "USD"
```

## Provider-Specific Features

### OpenAI Features
- Streaming responses
- Function calling
- Vision models
- Fine-tuned models

### Gemini Features
- Multimodal inputs (text, images)
- Function calling
- Grounding with Google Search
- Code execution

### Claude Features
- Large context windows (200K tokens)
- Advanced reasoning
- Constitutional AI safety

## Pricing Integration

Each provider includes real-time pricing information:

- **Input Token Cost:** Cost per 1,000 input tokens
- **Output Token Cost:** Cost per 1,000 output tokens
- **Currency:** Billing currency (typically USD)

Pricing data is used for cost optimization in load balancing.

## Rate Limiting

Each key has configurable rate limits:

- **Requests per Minute:** Maximum API calls per minute
- **Tokens per Minute:** Maximum tokens processed per minute

TruckLLM automatically rotates keys when limits are approached.

## Error Handling

Providers handle various error conditions:

- **Rate Limits (429):** Automatic retry with backoff
- **Authentication Errors (401/403):** Key rotation
- **Server Errors (5xx):** Failover to alternative providers
- **Model Not Found (404):** Fallback model selection

## Monitoring

Provider performance is tracked:

- Request latency
- Success/error rates
- Token usage
- Cost accumulation

Metrics are exposed via Prometheus and admin APIs.

## Best Practices

### Key Management
- Use multiple API keys per provider for redundancy
- Rotate keys regularly for security
- Monitor key usage and limits

### Cost Optimization
- Configure accurate pricing information
- Use cost-first strategies for budget-conscious deployments
- Monitor spending via admin APIs

### Reliability
- Configure multiple providers for failover
- Set appropriate rate limits
- Monitor error rates and latency

### Security
- Store API keys securely (environment variables)
- Use HTTPS for all provider communications
- Implement proper authentication and authorization