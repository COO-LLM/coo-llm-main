# Fireworks AI Provider

---
sidebar_position: 2
tags: [developer-guide, providers, fireworks]
---

# Fireworks AI Provider

Fireworks AI provides fast inference for open-source large language models, offering OpenAI-compatible APIs for seamless integration.

## Configuration

Add to your `llm_providers` array in `config.yaml`:

```yaml
llm_providers:
  - id: "fireworks"  # Provider ID for routing
    type: "fireworks"
    api_keys: ["${FIREWORKS_KEY}"]  # ENV var resolved at runtime
    base_url: "https://api.fireworks.ai/inference/v1"  # Optional
    model: "accounts/fireworks/models/llama-v3-70b-instruct"  # Default model
    pricing:
      input_token_cost: 0.0000009  # $0.90 per million tokens
      output_token_cost: 0.0000009  # $0.90 per million tokens
    limits:
      req_per_min: 100
      tokens_per_min: 100000
      max_tokens: 4096
      session_limit: 10000000
      session_type: "1d"
```

## API Key Setup

1. Go to [Fireworks AI Console](https://app.fireworks.ai/)
2. Create a new API key
3. Set as environment variable: `export FIREWORKS_KEY="your-key-here"`
4. Use `${FIREWORKS_KEY}` in config for runtime resolution
    limits:
      req_per_min: 60
      tokens_per_min: 60000
```

## Supported Models

### Chat Completion Models
- `accounts/fireworks/models/llama-v3-8b-instruct` - Llama 3 8B Instruct
- `accounts/fireworks/models/llama-v3-70b-instruct` - Llama 3 70B Instruct
- `accounts/fireworks/models/mixtral-8x7b-instruct` - Mixtral 8x7B Instruct
- `accounts/fireworks/models/qwen2-72b-instruct` - Qwen2 72B Instruct
- `accounts/fireworks/models/yi-large` - Yi Large
- `accounts/fireworks/models/gemma-7b-it` - Gemma 7B IT
- `accounts/fireworks/models/llama-v2-7b-chat` - Llama 2 7B Chat
- `accounts/fireworks/models/llama-v2-13b-chat` - Llama 2 13B Chat
- `accounts/fireworks/models/llama-v2-70b-chat` - Llama 2 70B Chat

### Embedding Models
Fireworks supports embeddings through their OpenAI-compatible API. Common embedding models include:
- `nomic-ai/nomic-embed-text-v1.5`
- `thenlper/gte-large`

## API Compatibility

- ✅ Chat Completions (`/v1/chat/completions`)
- ✅ Streaming responses
- ✅ Embeddings (`/v1/embeddings`)
- ✅ Model listing
- ✅ Function calling (where supported by underlying models)

## Features

- **High Performance**: Optimized inference for low latency
- **Open-Source Focus**: Access to popular open-source models
- **OpenAI Compatibility**: Drop-in replacement for OpenAI API
- **Cost Effective**: Competitive pricing for high-volume usage

## Rate Limits

Default limits (configurable):
- 60 requests per minute
- 60,000 tokens per minute

## Usage Examples

```bash
# Chat completion
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "fireworks:accounts/fireworks/models/llama-v3-8b-instruct",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'

# Embeddings
curl -X POST http://localhost:8080/v1/embeddings \
  -H "Authorization: Bearer your-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "fireworks:nomic-ai/nomic-embed-text-v1.5",
    "input": ["Hello world"]
  }'
```

## Getting API Keys

1. Sign up at [Fireworks AI](https://fireworks.ai/)
2. Generate an API key from your dashboard
3. Add the key to your COO-LLM configuration

## Notes

- Fireworks AI specializes in fast inference for open-source models
- Model names follow the `accounts/fireworks/models/{model-name}` format
- Pricing is typically lower than proprietary model providers
- Excellent choice for cost-conscious deployments requiring high throughput