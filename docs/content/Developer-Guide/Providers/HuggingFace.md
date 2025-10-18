---
sidebar_position: 9
tags: [developer-guide, providers, huggingface]
---

# Hugging Face Provider

Implementation details for Hugging Face Inference API provider integration.

## Configuration

```yaml
llm_providers:
  - id: "huggingface"
    type: "huggingface"
    api_keys: ["your-huggingface-token"]
    base_url: "https://api-inference.huggingface.co/v1"  # Optional
    model: "microsoft/DialoGPT-medium"
    pricing:
      input_token_cost: 0.0000002
      output_token_cost: 0.0000002
    limits:
      req_per_min: 300
      tokens_per_min: 100000
```

## Supported Models

Hugging Face hosts thousands of models. Popular ones include:

### Text Generation
- **DialoGPT**: `microsoft/DialoGPT-medium`, `microsoft/DialoGPT-large`
- **FLAN-T5**: `google/flan-t5-base`
- **Phi-2**: `microsoft/Phi-2`
- **Gemma**: `google/gemma-7b`

### Large Language Models
- **Mistral**: `mistralai/Mistral-7B-Instruct-v0.1`
- **Llama 2**: `meta-llama/Llama-2-7b-chat-hf`
- **BLOOM**: `bigscience/bloom-560m`

### Instruction-tuned Models
- **Phi-3**: `microsoft/Phi-3-mini-4k-instruct`
- **Llama 3.1**: `meta-llama/Llama-3.1-8B-Instruct`

## Embeddings Support

Hugging Face supports embeddings through compatible models:

- **Sentence Transformers**: Various embedding models
- **Input**: Text strings up to model limits
- **Output**: Variable dimensions depending on model

## API Mapping

### Chat Completions

| COO-LLM Parameter | Hugging Face API Parameter |
|-------------------|-----------------------------|
| `max_tokens` | `max_tokens` |
| `temperature` | `temperature` |
| `top_p` | `top_p` |
| `stream` | `stream` |
| `stop` | `stop` |

### Embeddings

| COO-LLM Parameter | Hugging Face API Parameter |
|-------------------|-----------------------------|
| `model` | `model` (embedding model) |
| `input` | `input` (text strings) |
| `user` | `user` (optional) |

## Authentication

Uses HTTP header:
```
Authorization: Bearer your-huggingface-token
```

## Error Handling

Hugging Face-specific errors:

- **429 Rate Limited**: Request too fast (free tier limits)
- **401 Unauthorized**: Invalid token
- **400 Bad Request**: Invalid parameters
- **503 Service Unavailable**: Model loading or server busy

## Rate Limiting

- **Free Tier**: 30,000 requests/month, 300 requests/minute
- **Paid Tier**: Higher limits available
- **Model Loading**: First request may be slower due to model loading

## Cost Calculation

Pricing varies by model and usage:

```go
inputCost := tokens.Input * 0.0000002 / 1000  // Example rate
outputCost := tokens.Output * 0.0000002 / 1000
totalCost := inputCost + outputCost
```

## Implementation Notes

- Uses OpenAI-compatible Inference API
- Supports streaming via server-sent events
- Extensive model catalog from open-source community
- Free tier available for experimentation
- Model loading time may affect first request latency
- Token counting estimated from input/output length