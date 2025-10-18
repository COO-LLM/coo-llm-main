---
sidebar_position: 1
tags: [developer-guide, providers, openai]
---

# OpenAI Provider

Implementation details for OpenAI provider integration.

## Configuration

```yaml
llm_providers:
  - id: "openai"
    type: "openai"
    api_keys: ["${OPENAI_KEY1}", "${OPENAI_KEY2}"]  # ENV vars resolved at runtime
    base_url: "https://api.openai.com/v1"  # Optional
    model: "gpt-4o"
    pricing:
      input_token_cost: 0.0025
      output_token_cost: 0.01
    limits:
      req_per_min: 200
      tokens_per_min: 100000
      max_tokens: 4096
      session_limit: 10000000
      session_type: "1d"
```

## API Key Setup

1. Go to [OpenAI API Keys](https://platform.openai.com/api-keys)
2. Create a new API key
3. Set as environment variable: `export OPENAI_KEY1="sk-your-key-here"`
4. Use `${OPENAI_KEY1}` in config for runtime resolution

## Supported Models

- **GPT-4o**: Most advanced model, best performance
- **GPT-4-turbo**: Fast, cost-effective
- **GPT-3.5-turbo**: Legacy, budget option

## API Mapping

| COO-LLM Parameter | OpenAI API Parameter |
|-------------------|----------------------|
| `max_tokens` | `max_tokens` |
| `temperature` | `temperature` |
| `top_p` | `top_p` |
| `stream` | `stream` |
| `stop` | `stop` |

## Error Handling

OpenAI-specific errors:

- **429 Rate Limited**: Automatic key rotation
- **401 Invalid Key**: Mark key as invalid
- **400 Bad Request**: Validate parameters
- **500 Server Error**: Retry with backoff

## Rate Limiting

- **Requests**: 200/min per key (varies by model)
- **Tokens**: 100,000/min per key
- **Organization limits**: May apply across all keys

## Cost Calculation

```go
inputCost := tokens.Input * 0.002 / 1000
outputCost := tokens.Output * 0.01 / 1000
totalCost := inputCost + outputCost
```

## Implementation Notes

- Uses REST API with Bearer authentication
- Supports streaming responses
- Handles token counting from response
- Automatic retry on transient errors