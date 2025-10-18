---
sidebar_position: 5
tags: [developer-guide, providers, together]
---

# Together AI Provider

Implementation details for Together AI provider integration.

## Configuration

```yaml
llm_providers:
  - id: "together"
    type: "together"
    api_keys: ["your-together-key"]
    base_url: "https://api.together.xyz/v1"  # Optional
    model: "meta-llama/Llama-3.3-70B-Instruct-Turbo"
    pricing:
      input_token_cost: 0.0000009
      output_token_cost: 0.0000009
    limits:
      req_per_min: 1000
      tokens_per_min: 1000000
```

## Supported Models

Together AI supports a wide range of models from various providers:

### Meta Llama Series
- **Llama 3.3**: `meta-llama/Llama-3.3-70B-Instruct-Turbo`
- **Llama 3.2**: `meta-llama/Llama-3.2-3B-Instruct-Turbo`, `meta-llama/Llama-3.2-11B-Vision-Instruct-Turbo`
- **Llama 3**: `meta-llama/Llama-3-8B-Instruct-Turbo`, `meta-llama/Llama-3-70B-Instruct-Turbo`

### Mistral AI
- **Mistral**: `mistralai/Mistral-7B-Instruct-v0.1`
- **Mixtral**: `mistralai/Mixtral-8x7B-Instruct-v0.1`

### Google
- **Gemma**: `google/gemma-2-9b-it`, `google/gemma-2-27b-it`

### Microsoft
- **WizardLM**: `microsoft/WizardLM-2-8x22B`

### Databricks
- **DBRX**: `databricks/dbrx-instruct`

### Alibaba
- **Qwen**: `Qwen/Qwen2.5-72B-Instruct-Turbo`, `Qwen/Qwen2.5-Coder-32B-Instruct`

## Embeddings Support

Together AI supports embeddings through OpenAI-compatible API:

- **Embedding Models**: Compatible with OpenAI embedding models
- **Input**: Text strings up to model limits
- **Output**: Variable dimensions depending on model
- **Batch Processing**: Support for multiple inputs

## API Mapping

### Chat Completions

| COO-LLM Parameter | Together API Parameter |
|-------------------|------------------------|
| `max_tokens` | `max_tokens` |
| `temperature` | `temperature` |
| `top_p` | `top_p` |
| `stream` | `stream` |
| `stop` | `stop` |

### Embeddings

| COO-LLM Parameter | Together API Parameter |
|-------------------|-------------------------|
| `model` | `model` (embedding model) |
| `input` | `input` (text strings) |
| `user` | `user` (optional) |

## Authentication

Uses HTTP header:
```
Authorization: Bearer your-together-key
```

## Error Handling

Together AI-specific errors:

- **429 Rate Limited**: Request too fast
- **401 Unauthorized**: Invalid API key
- **400 Bad Request**: Invalid parameters
- **500 Internal Error**: Server error
- **503 Service Unavailable**: Model temporarily unavailable

## Rate Limiting

- **Requests**: 1000/min per key
- **Tokens**: 1,000,000/min per key
- **Higher limits** available for paid plans

## Cost Calculation

Pricing varies by model, example for Llama 3.3 70B:

```go
inputCost := tokens.Input * 0.0000009 / 1000
outputCost := tokens.Output * 0.0000009 / 1000
totalCost := inputCost + outputCost
```

## Implementation Notes

- Uses OpenAI-compatible API endpoints
- Supports streaming via server-sent events
- Extensive model catalog from multiple providers
- High rate limits suitable for production use
- Token counting from response metadata