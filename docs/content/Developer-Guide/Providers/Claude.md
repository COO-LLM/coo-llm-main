---
sidebar_position: 3
tags: [developer-guide, providers, claude]
---

# Claude Provider

Implementation details for Anthropic Claude provider integration.

## Configuration

```yaml
llm_providers:
  - id: "claude"
    type: "claude"
    api_keys: ["${CLAUDE_KEY1}", "${CLAUDE_KEY2}"]  # ENV vars resolved at runtime
    base_url: "https://api.anthropic.com"  # Optional
    model: "claude-3-5-sonnet-20240620"
    pricing:
      input_token_cost: 0.003
      output_token_cost: 0.015
    limits:
      req_per_min: 100
      tokens_per_min: 80000
      max_tokens: 4096
      session_limit: 10000000
      session_type: "1d"
```

## API Key Setup

1. Go to [Anthropic Console](https://console.anthropic.com/)
2. Create a new API key
3. Set as environment variable: `export CLAUDE_KEY1="sk-ant-your-key-here"`
4. Use `${CLAUDE_KEY1}` in config for runtime resolution

## Supported Models

- **Claude 3 Opus**: Most capable model
- **Claude 3 Sonnet**: Balanced performance/cost
- **Claude 3 Haiku**: Fast, efficient
- **Claude 3.5 Sonnet**: Latest high-performance

## Embeddings Support

Claude does not provide native embedding models. Anthropic recommends using **Voyage AI** for embeddings:

- **Recommended Models**: `voyage-3-large`, `voyage-3.5`, `voyage-3.5-lite`
- **Specialized Models**: `voyage-code-3` (code), `voyage-finance-2` (finance), `voyage-law-2` (legal)
- **Multimodal**: `voyage-multimodal-3` for text + images

For embedding functionality, consider using Voyage AI as a separate provider in COO-LLM.

## API Mapping

| COO-LLM Parameter | Claude API Parameter |
|-------------------|----------------------|
| `max_tokens` | `max_tokens` |
| `temperature` | `temperature` |
| `top_p` | `top_p` |
| `stream` | `stream` |
| `stop` | `stop_sequences` |

## Authentication

Uses HTTP header:
```
x-api-key: sk-ant-your-key
anthropic-version: 2023-06-01
```

## Error Handling

Claude-specific errors:

- **429 Rate Limited**: Request too fast
- **401 Unauthorized**: Invalid API key
- **400 Bad Request**: Invalid parameters
- **529 Overloaded**: Server overloaded

## Rate Limiting

- **Requests**: 100/min per key
- **Tokens**: 60,000/min per key
- **Input tokens**: 20,000/min per key
- **Output tokens**: 40,000/min per key

## Cost Calculation

```go
inputCost := tokens.Input * 0.015 / 1000
outputCost := tokens.Output * 0.075 / 1000
totalCost := inputCost + outputCost
```

## Implementation Notes

- Uses HTTP POST with custom headers
- Supports streaming via event streams
- Token counting from usage field
- Handles Anthropic-specific error formats