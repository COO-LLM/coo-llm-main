---
sidebar_position: 10
tags: [developer-guide, providers, replicate]
---

# Replicate Provider

Implementation details for Replicate model hosting provider integration.

## Configuration

```yaml
llm_providers:
  - id: "replicate"
    type: "replicate"
    api_keys: ["your-replicate-token"]
    base_url: "https://api.replicate.com"  # Optional
    model: "meta/llama-2-70b-chat"
    pricing:
      input_token_cost: 0.00000065
      output_token_cost: 0.0000026
    limits:
      req_per_min: 100
      tokens_per_min: 10000
```

## Supported Models

Replicate hosts a wide variety of models from different providers:

### Meta Llama
- **Llama 2 70B Chat**: `meta/llama-2-70b-chat`
- **Llama 2 7B Chat**: `replicate/llama-2-70b-chat`

### Mistral AI
- **Mistral 7B Instruct**: `mistralai/mistral-7b-instruct-v0.1`

### Stability AI
- **Stable Diffusion**: `stability-ai/stable-diffusion`

### Other Models
- **Whisper**: `openai/whisper` (speech-to-text)
- **MiniGPT-4**: `daanelson/minigpt-4` (VQA)
- **AnimateDiff**: `lucataco/animate-diff` (animation)

## Embeddings Support

Replicate does not have dedicated embedding models. Most hosted models are for text generation or other tasks.

- **Recommendation**: Use Cohere, Voyage AI, or OpenAI for embeddings

## API Mapping

### Chat Completions

| COO-LLM Parameter | Replicate API Parameter |
|-------------------|--------------------------|
| `max_tokens` | `max_tokens` |
| `temperature` | `temperature` |
| `top_p` | `top_p` |
| `stream` | Not supported |
| `stop` | `stop_sequences` |

### Embeddings

| COO-LLM Parameter | Replicate API Parameter |
|-------------------|--------------------------|
| `model` | N/A |
| `input` | N/A |
| `user` | N/A |

## Authentication

Uses HTTP header:
```
Authorization: Token your-replicate-token
```

## Error Handling

Replicate-specific errors:

- **429 Rate Limited**: Request too fast
- **401 Unauthorized**: Invalid token
- **400 Bad Request**: Invalid parameters
- **402 Payment Required**: Insufficient credits
- **500 Internal Error**: Server error

## Rate Limiting

- **Free Tier**: Limited credits
- **Paid Tier**: Based on model usage and credits purchased
- **Model Loading**: First request may be slower

## Cost Calculation

Pricing varies by model and compute time:

```go
// Example for Llama 2 70B
inputCost := tokens.Input * 0.00000065 / 1000
outputCost := tokens.Output * 0.0000026 / 1000
totalCost := inputCost + outputCost
```

## Implementation Notes

- Uses REST API with JSON payloads
- Synchronous predictions (no streaming support)
- Model versioning system (version IDs)
- Pay-per-use pricing model
- Token counting estimated from input/output length
- May require polling for long-running predictions