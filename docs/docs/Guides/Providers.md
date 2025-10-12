---
sidebar_position: 4
tags: [user-guide, providers]
---

# Providers

COO-LLM supports multiple LLM providers through a plugin-based architecture. Each provider implements a common interface for seamless integration.

## Supported Providers

### OpenAI

**Provider Type:** `openai`

**Configuration:**
```yaml
llm_providers:
  - id: "openai-prod"
    type: "openai"
    api_keys: ["${OPENAI_KEY_1}", "${OPENAI_KEY_2}"]
    base_url: "https://api.openai.com"
    model: "gpt-4o"
    pricing:
      input_token_cost: 0.002
      output_token_cost: 0.01
    limits:
      req_per_min: 200
      tokens_per_min: 100000
```

**Supported Models:**
- `gpt-4`
- `gpt-4-turbo`
- `gpt-4o`
- `gpt-3.5-turbo`
- `gpt-3.5-turbo-instruct`

**Features:**
- Chat completions with conversation history
- Token usage tracking
- Error handling and retries
- Rate limit management

**Rate Limits:** Based on OpenAI tier (see [OpenAI docs](https://platform.openai.com/docs/guides/rate-limits))

### Google Gemini

**Provider Type:** `gemini`

**Configuration:**
```yaml
llm_providers:
  - id: "gemini-prod"
    type: "gemini"
    api_keys: ["${GEMINI_KEY_1}"]
    base_url: "https://generativelanguage.googleapis.com"
    model: "gemini-1.5-pro"
    pricing:
      input_token_cost: 0.00025
      output_token_cost: 0.0005
    limits:
      req_per_min: 150
      tokens_per_min: 80000
```

**Supported Models:**
- `gemini-1.5-pro`
- `gemini-1.5-flash`
- `gemini-pro`

**Features:**
- Chat completions
- Token usage tracking
- Error handling

**Rate Limits:** Varies by tier (see [Gemini docs](https://ai.google.dev/docs/rate-limits))

### Anthropic Claude

**Provider Type:** `claude`

**Configuration:**
```yaml
llm_providers:
  - id: "claude-prod"
    type: "claude"
    api_keys: ["${CLAUDE_KEY_1}"]
    base_url: "https://api.anthropic.com"
    model: "claude-3-opus"
    pricing:
      input_token_cost: 0.015
      output_token_cost: 0.075
    limits:
      req_per_min: 100
      tokens_per_min: 60000
```

**Supported Models:**
- `claude-3-opus`
- `claude-3-sonnet`
- `claude-3-haiku`
- `claude-2.1`

**Features:**
- Chat completions
- Token usage tracking
- Error handling

**Rate Limits:** Varies by model (see [Anthropic docs](https://docs.anthropic.com/claude/docs/rate-limits))

### Custom Provider

**Provider Type:** `custom`

**Configuration:**
```yaml
llm_providers:
  - id: "custom-dev"
    type: "custom"
    api_keys: ["${CUSTOM_KEY_1}"]
    base_url: "https://custom-llm.example.com"
    model: "custom-model"
    pricing:
      input_token_cost: 0.001
      output_token_cost: 0.002
    limits:
      req_per_min: 50
      tokens_per_min: 10000
```

**Use Case:** Integrate with proprietary or custom LLM APIs

## Provider Interface

All providers implement the `LLMProvider` interface defined in `internal/provider/interface.go`:

```go
type LLMProvider interface {
    Name() string
    Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error)
    ListModels(ctx context.Context) ([]string, error)
}
```

### Request Structure

```go
type LLMRequest struct {
    Prompt    string                   `json:"prompt"`
    Messages  []map[string]interface{} `json:"messages,omitempty"`
    Model     string                   `json:"model,omitempty"`
    MaxTokens int                      `json:"max_tokens,omitempty"`
    Params    map[string]any           `json:"params,omitempty"`
}
```

### Response Structure

```go
type LLMResponse struct {
    Text         string `json:"text"`
    InputTokens  int    `json:"input_tokens"`
    OutputTokens int    `json:"output_tokens"`
    TokensUsed   int    `json:"tokens_used"`
    FinishReason string `json:"finish_reason"`
}
```

### Configuration Structure

```go
type LLMConfig struct {
    Type    ProviderType `yaml:"type"`
    APIKeys []string     `yaml:"api_keys"`
    BaseURL string       `yaml:"base_url,omitempty"`
    Model   string       `yaml:"model"`
    Pricing Pricing      `yaml:"pricing"`
    Limits  Limits       `yaml:"limits"`
}
```

## Adding Custom Providers

To add a new provider, implement the `LLMProvider` interface and register it in the registry.

1. **Create a new provider file (e.g., `internal/provider/custom.go`):**

```go
package provider

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type CustomProvider struct {
    config LLMConfig
}

func NewCustomProvider(config LLMConfig) LLMProvider {
    return &CustomProvider{config: config}
}

func (p *CustomProvider) Name() string {
    return "custom" // Provider type identifier
}

func (p *CustomProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
    // Get API key (round-robin if multiple)
    apiKey := p.config.APIKey()
    if apiKey == "" {
        return nil, fmt.Errorf("no API key available")
    }

    // Transform to provider-specific request format
    providerReq := map[string]interface{}{
        "model": req.Model,
        "messages": req.Messages,
        "max_tokens": req.MaxTokens,
    }

    // Make HTTP request
    jsonData, _ := json.Marshal(providerReq)
    httpReq, err := http.NewRequestWithContext(ctx, "POST",
        p.config.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+apiKey)

    client := &http.Client{Timeout: 30 * time.Second}
    resp, err := client.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Read response
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("provider error: %s", string(body))
    }

    // Parse provider response
    var providerResp map[string]interface{}
    if err := json.Unmarshal(body, &providerResp); err != nil {
        return nil, err
    }

    // Extract response data
    choices := providerResp["choices"].([]interface{})
    if len(choices) == 0 {
        return nil, fmt.Errorf("no choices in response")
    }

    choice := choices[0].(map[string]interface{})
    message := choice["message"].(map[string]interface{})
    text := message["content"].(string)

    // Extract usage (adjust based on provider's response format)
    usage := providerResp["usage"].(map[string]interface{})
    inputTokens := int(usage["prompt_tokens"].(float64))
    outputTokens := int(usage["completion_tokens"].(float64))

    return &LLMResponse{
        Text:         text,
        InputTokens:  inputTokens,
        OutputTokens: outputTokens,
        TokensUsed:   inputTokens + outputTokens,
        FinishReason: choice["finish_reason"].(string),
    }, nil
}

func (p *CustomProvider) ListModels(ctx context.Context) ([]string, error) {
    // Return supported models for this provider
    return []string{p.config.Model}, nil
}
```

2. **Register the provider in `internal/provider/registry.go`:**

```go
func (r *Registry) LoadFromConfig(cfg *config.Config) error {
    // ... existing code for LLMProviders ...

    for _, lp := range cfg.LLMProviders {
        llmCfg := LLMConfig{
            Type:    ProviderType(lp.Type),
            APIKeys: lp.APIKeys,
            BaseURL: lp.BaseURL,
            Model:   lp.Model,
            Pricing: lp.Pricing,
            Limits:  lp.Limits,
        }
        var p LLMProvider
        switch lp.Type {
        case "openai":
            p = NewOpenAIProvider(llmCfg)
        case "gemini":
            p = NewGeminiProvider(llmCfg)
        case "claude":
            p = NewClaudeProvider(llmCfg)
        case "custom":
            p = NewCustomProvider(llmCfg)
        default:
            return fmt.Errorf("unsupported provider type: %s", lp.Type)
        }
        r.Register(p)
    }
    return nil
}
```

3. **Add to configuration:**

```yaml
llm_providers:
  - id: "my-custom"
    type: "custom"
    api_keys: ["${CUSTOM_API_KEY}"]
    base_url: "https://api.my-custom-provider.com/v1"
    model: "my-model"
    pricing:
      input_token_cost: 0.001
      output_token_cost: 0.002
    limits:
      req_per_min: 100
      tokens_per_min: 50000

model_aliases:
  my-model: my-custom:my-model
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

COO-LLM automatically rotates keys when limits are approached.

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