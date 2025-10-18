---
sidebar_position: 6
tags: [developer-guide, providers, openrouter]
---

# OpenRouter Provider

Implementation details for OpenRouter provider integration.

## Configuration

```yaml
llm_providers:
  - id: "openrouter"
    type: "openrouter"
    api_keys: ["sk-or-v1-your-key"]
    base_url: "https://openrouter.ai/api/v1"  # Optional
    model: "anthropic/claude-3.5-sonnet"
    pricing:
      input_token_cost: 0.003
      output_token_cost: 0.015
    limits:
      req_per_min: 100
      tokens_per_min: 10000
```

## Supported Models

OpenRouter provides access to hundreds of models from various providers:

### Anthropic
- **Claude 3.5 Sonnet**: `anthropic/claude-3.5-sonnet`
- **Claude 3 Opus**: `anthropic/claude-3-opus`
- **Claude 3 Haiku**: `anthropic/claude-3-haiku`

### OpenAI
- **GPT-4o**: `openai/gpt-4o`
- **GPT-4o Mini**: `openai/gpt-4o-mini`
- **GPT-3.5 Turbo**: `openai/gpt-3.5-turbo`

### Meta
- **Llama 3.2 90B**: `meta-llama/llama-3.2-90b-instruct`
- **Llama 3.1 70B**: `meta-llama/llama-3.1-70b-instruct`

### Google
- **Gemini Pro 1.5**: `google/gemini-pro-1.5`

### Mistral
- **Mistral 7B**: `mistralai/mistral-7b-instruct`

And many more models from various providers.

## Embeddings Support

OpenRouter supports embeddings through OpenAI-compatible API:

- **Compatible with OpenAI embedding models**
- **Input**: Text strings up to model limits
- **Output**: Variable dimensions depending on model

## API Mapping

### Chat Completions

| COO-LLM Parameter | OpenRouter API Parameter |
|-------------------|--------------------------|
| `max_tokens` | `max_tokens` |
| `temperature` | `temperature` |
| `top_p` | `top_p` |
| `stream` | `stream` |
| `stop` | `stop` |

### Embeddings

| COO-LLM Parameter | OpenRouter API Parameter |
|-------------------|---------------------------|
| `model` | `model` (embedding model) |
| `input` | `input` (text strings) |
| `user` | `user` (optional) |

## Authentication

Uses HTTP header:
```
Authorization: Bearer sk-or-v1-your-key
```

## Error Handling

OpenRouter-specific errors:

- **429 Rate Limited**: Request too fast
- **401 Unauthorized**: Invalid API key
- **400 Bad Request**: Invalid parameters
- **402 Payment Required**: Insufficient credits
- **503 Service Unavailable**: Model temporarily unavailable

## Rate Limiting

- **Requests**: 100/min per key (varies by plan)
- **Tokens**: 10,000/min per key (varies by plan)
- **Higher limits** available for paid plans

## Cost Calculation

Pricing varies by model and provider. Example for Claude 3.5 Sonnet:

```go
inputCost := tokens.Input * 0.003 / 1000
outputCost := tokens.Output * 0.015 / 1000
totalCost := inputCost + outputCost
```

## Implementation Notes

- Uses OpenAI-compatible API endpoints
- Supports streaming via server-sent events
- Extensive model catalog from multiple providers
- Pay-per-use pricing model
- Token counting from response metadata
- Automatic provider routing