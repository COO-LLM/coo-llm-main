---
sidebar_position: 8
tags: [developer-guide, providers, cohere]
---

# Cohere Provider

Implementation details for Cohere provider integration.

## Configuration

```yaml
llm_providers:
  - id: "cohere"
    type: "cohere"
    api_keys: ["your-cohere-key"]
    base_url: "https://api.cohere.ai"  # Optional
    model: "command-r-plus"
    pricing:
      input_token_cost: 0.0015
      output_token_cost: 0.006
    limits:
      req_per_min: 100
      tokens_per_min: 10000
```

## Supported Models

### Command Series (Chat)
- **Command R+**: `command-r-plus` - Most capable model
- **Command R**: `command-r` - Balanced performance
- **Command**: `command` - Previous generation
- **Command Light**: `command-light` - Fast and efficient

### Embedding Models
- **Embed English v3.0**: `embed-english-v3.0` - English embeddings
- **Embed Multilingual v3.0**: `embed-multilingual-v3.0` - Multi-language support
- **Embed English Light v3.0**: `embed-english-light-v3.0` - Fast English embeddings
- **Embed Multilingual Light v3.0**: `embed-multilingual-light-v3.0` - Fast multi-language

## Embeddings Support

Cohere has excellent embeddings support:

- **English Models**: `embed-english-v3.0`, `embed-english-light-v3.0`
- **Multilingual Models**: `embed-multilingual-v3.0`, `embed-multilingual-light-v3.0`
- **Dimensions**: 1024 (default), 384, 256, 512, 2048
- **Input**: Text up to 512 tokens per input
- **Batch Processing**: Support for multiple inputs

## API Mapping

### Chat Completions

| COO-LLM Parameter | Cohere API Parameter |
|-------------------|----------------------|
| `max_tokens` | `max_tokens` |
| `temperature` | `temperature` |
| `top_p` | `p` |
| `stream` | `stream` |
| `stop` | `stop_sequences` |

### Embeddings

| COO-LLM Parameter | Cohere API Parameter |
|-------------------|----------------------|
| `model` | `model` (embedding model) |
| `input` | `texts` (array of strings) |
| `user` | Not supported |

## Authentication

Uses HTTP header:
```
Authorization: Bearer your-cohere-key
```

## Error Handling

Cohere-specific errors:

- **429 Rate Limited**: Request too fast
- **401 Unauthorized**: Invalid API key
- **400 Bad Request**: Invalid parameters
- **500 Internal Error**: Server error

## Rate Limiting

- **Requests**: 100/min per key
- **Tokens**: 10,000/min per key
- **Higher limits** available for enterprise plans

## Cost Calculation

Pricing varies by model. Example for Command R+:

```go
inputCost := tokens.Input * 0.0015 / 1000
outputCost := tokens.Output * 0.006 / 1000
totalCost := inputCost + outputCost
```

## Implementation Notes

- Uses REST API with JSON payloads
- Supports streaming via server-sent events
- Excellent embeddings quality and performance
- Token counting estimated from input/output length
- Handles Cohere-specific error formats
- Strong multilingual support