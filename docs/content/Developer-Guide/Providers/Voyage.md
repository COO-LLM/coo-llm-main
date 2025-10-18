---
sidebar_position: 11
tags: [developer-guide, providers, voyage]
---

# Voyage AI Provider

Implementation details for Voyage AI embeddings provider integration.

## Configuration

```yaml
llm_providers:
  - id: "voyage"
    type: "voyage"
    api_keys: ["your-voyage-key"]
    base_url: "https://api.voyageai.com"  # Optional
    model: "voyage-3-large"
    pricing:
      input_token_cost: 0.00000012
      output_token_cost: 0.00000012
    limits:
      req_per_min: 1000
      tokens_per_min: 1000000
```

## Supported Models

Voyage AI specializes in high-quality text embeddings:

### General Purpose
- **Voyage 3 Large**: `voyage-3-large` - Best quality, 1024 dimensions
- **Voyage 3.5**: `voyage-3.5` - Balanced performance, 1024 dimensions
- **Voyage 3.5 Lite**: `voyage-3.5-lite` - Fast and cost-effective, 1024 dimensions

### Domain-Specific
- **Voyage Code 3**: `voyage-code-3` - Optimized for code retrieval
- **Voyage Finance 2**: `voyage-finance-2` - Specialized for finance
- **Voyage Law 2**: `voyage-law-2` - Optimized for legal documents

### Multimodal
- **Voyage Multimodal 3**: `voyage-multimodal-3` - Text and image embeddings

## Embeddings Support

Voyage AI is specialized in embeddings with state-of-the-art quality:

- **Dimensions**: 1024 (default), supports 256, 512, 2048
- **Input**: Text strings up to 512 tokens
- **Batch Processing**: Support for multiple inputs
- **Input Types**: `document`, `query` for different use cases

## API Mapping

### Chat Completions

| COO-LLM Parameter | Voyage API Parameter |
|-------------------|----------------------|
| `max_tokens` | N/A |
| `temperature` | N/A |
| `top_p` | N/A |
| `stream` | N/A |
| `stop` | N/A |

*Note: Voyage AI only supports embeddings, not text generation*

### Embeddings

| COO-LLM Parameter | Voyage API Parameter |
|-------------------|----------------------|
| `model` | `model` (embedding model) |
| `input` | `input` (array of strings) |
| `user` | Not supported |

## Authentication

Uses HTTP header:
```
Authorization: Bearer your-voyage-key
```

## Error Handling

Voyage AI-specific errors:

- **429 Rate Limited**: Request too fast
- **401 Unauthorized**: Invalid API key
- **400 Bad Request**: Invalid parameters
- **500 Internal Error**: Server error

## Rate Limiting

- **Requests**: 1000/min per key
- **Tokens**: 1,000,000/min per key
- **Higher limits** available for enterprise plans

## Cost Calculation

Pricing based on tokens processed:

```go
// All models have the same pricing
cost := totalTokens * 0.00000012 / 1000
```

## Implementation Notes

- Uses REST API with JSON payloads
- Specialized in embeddings only (no text generation)
- State-of-the-art embedding quality
- Supports batch processing efficiently
- Token counting from API response usage field
- Excellent for RAG and semantic search applications