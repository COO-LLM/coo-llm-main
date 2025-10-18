---
sidebar_position: 4
tags: [developer-guide, providers, grok]
---

# Grok Provider

Implementation details for xAI Grok provider integration.

## Configuration

```yaml
llm_providers:
  - id: "grok"
    type: "grok"
    api_keys: ["xai-your-key"]
    base_url: "https://api.x.ai/v1"  # Optional
    model: "grok-beta"
    pricing:
      input_token_cost: 0.005
      output_token_cost: 0.015
    limits:
      req_per_min: 30
      tokens_per_min: 15000
```

## Supported Models

- **Grok Beta**: Latest Grok model
- **Grok Vision Beta**: Model with vision capabilities

## Embeddings Support

Grok supports text embeddings through OpenAI-compatible API:

- **Embedding Models**: Compatible with OpenAI embedding models
- **Input**: Text strings up to model limits
- **Output**: 1536-dimensional vectors (varies by model)

## API Mapping

### Chat Completions

| COO-LLM Parameter | Grok API Parameter |
|-------------------|---------------------|
| `max_tokens` | `max_tokens` |
| `temperature` | `temperature` |
| `top_p` | `top_p` |
| `stream` | `stream` |
| `stop` | `stop` |

### Embeddings

| COO-LLM Parameter | Grok API Parameter |
|-------------------|---------------------|
| `model` | `model` (embedding model) |
| `input` | `input` (text strings) |
| `user` | `user` (optional) |

## Authentication

Uses HTTP header:
```
Authorization: Bearer xai-your-key
```

## Error Handling

Grok-specific errors:

- **429 Rate Limited**: Request too fast
- **401 Unauthorized**: Invalid API key
- **400 Bad Request**: Invalid parameters
- **500 Internal Error**: Server error

## Rate Limiting

- **Requests**: 30/min per key
- **Tokens**: 15,000/min per key

## Cost Calculation

```go
inputCost := tokens.Input * 0.005 / 1000
outputCost := tokens.Output * 0.015 / 1000
totalCost := inputCost + outputCost
```

## Implementation Notes

- Uses OpenAI-compatible API endpoints
- Supports streaming via server-sent events
- Token counting from response metadata
- Handles xAI-specific error formats