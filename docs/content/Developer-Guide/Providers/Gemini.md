---
sidebar_position: 3
tags: [developer-guide, providers, gemini]
---

# Google Gemini Provider

Implementation details for Google Gemini provider integration.

## Configuration

```yaml
llm_providers:
  - id: "gemini-prod"
    type: "gemini"
    api_keys: ["${GEMINI_KEY1}", "${GEMINI_KEY2}"]  # ENV vars resolved at runtime
    model: "gemini-2.0-flash-exp"
    pricing:
      input_token_cost: 0.00025
      output_token_cost: 0.0005
    limits:
      req_per_min: 150
      tokens_per_min: 80000
      max_tokens: 10000000
      session_limit: 10000000
      session_type: "1d"
```

## API Key Setup

1. Go to [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Create a new API key
3. Set as environment variable: `export GEMINI_KEY="your-key-here"`
4. Use `${GEMINI_KEY}` in config for runtime resolution

## Supported Models

- **gemini-2.0-flash-exp**: Experimental 2.0 model (latest)
- **gemini-1.5-pro**: Advanced reasoning model
- **gemini-1.5-flash**: Fast, cost-effective model

## API Mapping

| COO-LLM Parameter | Gemini API Parameter |
|-------------------|----------------------|
| `max_tokens` | `MaxTokens` |
| `temperature` | `Temperature` |
| `top_p` | `TopP` |
| `stream` | Streaming supported |
| `messages` | Converted to Gemini format |

## Features

- **Multimodal**: Supports text, images, audio
- **Function Calling**: Advanced tool integration
- **Streaming**: Real-time response streaming
- **Safety Settings**: Configurable content filtering

## Error Handling

Gemini-specific errors:

- **API_KEY_INVALID**: Invalid or expired API key
- **RESOURCE_EXHAUSTED**: Rate limit exceeded
- **FAILED_PRECONDITION**: Model not available
- **INVALID_ARGUMENT**: Bad request parameters

## Implementation Notes

- Uses Google Generative AI Go SDK
- Automatic key rotation on failures
- Supports both Generate and GenerateStream methods
- Context timeout handling
- Error retry with exponential backoff
