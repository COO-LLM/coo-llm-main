---
sidebar_position: 7
tags: [developer-guide, providers, mistral]
---

# Mistral AI Provider

Implementation details for Mistral AI provider integration.

## Configuration

```yaml
llm_providers:
  - id: "mistral"
    type: "mistral"
    api_keys: ["your-mistral-key"]
    base_url: "https://api.mistral.ai"  # Optional
    model: "mistral-large-latest"
    pricing:
      input_token_cost: 0.002
      output_token_cost: 0.006
    limits:
      req_per_min: 100
      tokens_per_min: 10000
```

## Supported Models

- **Mistral Large**: `mistral-large-latest` - Most capable model
- **Mistral Medium**: `mistral-medium` - Balanced performance
- **Mistral Small**: `mistral-small` - Fast and efficient
- **Mistral 7B**: `mistral-7b-instruct` - Open source model
- **Mistral 8x7B**: `mistral-8x7b-instruct` - Mixture of experts
- **Mistral Embed**: `mistral-embed` - Embedding model

## Embeddings Support

Mistral supports text embeddings:

- **Embedding Model**: `mistral-embed`
- **Input**: Text strings up to 512 tokens
- **Output**: 1024-dimensional vectors
- **Context**: Up to 512 tokens per input

## API Mapping

### Chat Completions

| COO-LLM Parameter | Mistral API Parameter |
|-------------------|------------------------|
| `max_tokens` | `max_tokens` |
| `temperature` | `temperature` |
| `top_p` | `top_p` |
| `stream` | `stream` |
| `stop` | `stop_sequences` |

### Embeddings

| COO-LLM Parameter | Mistral API Parameter |
|-------------------|-----------------------|
| `model` | `model` (use `mistral-embed`) |
| `input` | `input` (text strings) |
| `user` | `user` (optional) |

## Authentication

Uses HTTP header:
```
Authorization: Bearer your-mistral-key
```

## Error Handling

Mistral-specific errors:

- **429 Rate Limited**: Request too fast
- **401 Unauthorized**: Invalid API key
- **400 Bad Request**: Invalid parameters
- **500 Internal Error**: Server error

## Rate Limiting

- **Requests**: 100/min per key
- **Tokens**: 10,000/min per key
- **Higher limits** available for enterprise plans

## Cost Calculation

```go
inputCost := tokens.Input * 0.002 / 1000
outputCost := tokens.Output * 0.006 / 1000
totalCost := inputCost + outputCost
```

## Implementation Notes

- Uses REST API with JSON payloads
- Supports streaming via server-sent events
- Token counting from usage field in responses
- Handles Mistral-specific error formats
- European data residency available