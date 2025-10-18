---
sidebar_position: 1
tags: [reference, api, llm]
---

# LLM API Reference

OpenAI-compatible REST APIs for LLM interactions.

## Authentication

All API requests require authentication via Bearer token:

```
Authorization: Bearer <api_key>
```

API keys are configured in the `api_keys` section and mapped to specific providers.

## Model Resolution

COO-LLM supports 3 ways to specify models:

1. **Direct provider:model syntax**: `provider_id:model_name` (e.g., `openai:gpt-4o`, `custom:my-model`)
2. **Model aliases**: Short names mapped to provider:model (e.g., `gpt-4o` → `openai:gpt-4o`)
3. **Pattern matching fallback**: Infer provider from model name (e.g., `gpt-4o` → OpenAI provider)

## Authentication & Authorization Flow

```mermaid
flowchart TD
    classDef client fill:#28a745,color:#fff,stroke:#fff,stroke-width:2px
    classDef auth fill:#ffc107,color:#000,stroke:#000,stroke-width:2px
    classDef process fill:#dc3545,color:#fff,stroke:#fff,stroke-width:2px
    classDef deny fill:#dc3545,color:#fff,stroke:#fff,stroke-width:2px

    A[Client Request<br/>with Bearer token]:::client
    A --> B[Extract API Key<br/>from Authorization header]:::auth
    B --> C{Key exists in<br/>api_keys config?}:::auth

    C -->|No| D[401 Unauthorized<br/>Invalid API key]:::deny
    C -->|Yes| E[Get allowed_providers<br/>for this key]:::auth
    E --> F{Provider allowed<br/>for this request?}:::auth

    F -->|No| G[403 Forbidden<br/>Access denied to provider]:::deny
    F -->|Yes| H[Proceed to<br/>Request Processing]:::process
```

## API Request Flow

```mermaid
flowchart TD
    classDef client fill:#28a745,color:#fff,stroke:#fff,stroke-width:2px
    classDef process fill:#dc3545,color:#fff,stroke:#fff,stroke-width:2px
    classDef external fill:#007bff,color:#fff,stroke:#fff,stroke-width:2px

    A[Client Request<br/>POST /v1/chat/completions]:::client
    A --> B[Authentication<br/>Bearer Token Validation]:::process
    B --> C[Model Alias Resolution<br/>Map to provider:model]:::process
    C --> D[Provider & Key Selection<br/>Load Balancing Algorithm]:::process
    D --> E[Rate Limit Check<br/>Per-key limits]:::process
    E --> F[External API Call<br/>OpenAI/Gemini/Claude]:::external
    F --> G[Response Processing<br/>Token counting, caching]:::process
    G --> H[Usage Tracking<br/>Metrics update]:::process
    H --> I[Return OpenAI-compatible<br/>JSON Response]:::client

    B --> J[401 Unauthorized]:::process
    E --> K[429 Rate Limited]:::process
    F --> L[Retry Logic<br/>Up to max_attempts]:::process
    L --> F
    L --> M[502/503 Error<br/>Provider failure]:::process
```

## OpenAI-Compatible Endpoints

### POST /v1/chat/completions

Generate chat completions using available models. This is the primary endpoint implemented in COO-LLM.

**Request Body:**
```json
{
  "model": "gpt-4o",
  "messages": [
    {
      "role": "user",
      "content": "Hello, how are you?"
    }
  ],
  "max_tokens": 100
}
```

**Response:**
```json
{
  "id": "chatcmpl-1234567890",
  "object": "chat.completion",
  "created": 1699123456,
  "model": "gpt-4o",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! I'm doing well, thank you for asking. How can I help you today?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 13,
    "completion_tokens": 17,
    "total_tokens": 30
  }
}
```

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `model` | string | Yes | - | Model identifier (provider:model or alias) |
| `messages` | array | Yes | - | Array of message objects |
| `max_tokens` | int | No | Provider default | Maximum tokens to generate |
| `temperature` | float | No | 1.0 | Sampling temperature (0.0-2.0) |
| `top_p` | float | No | 1.0 | Nucleus sampling (0.0-1.0) |
| `stream` | bool | No | false | Enable streaming response |
| `stop` | string/array | No | - | Stop sequences |
| `presence_penalty` | float | No | 0.0 | Presence penalty (-2.0-2.0) |
| `frequency_penalty` | float | No | 0.0 | Frequency penalty (-2.0-2.0) |
| `user` | string | No | - | A unique identifier representing your end-user. |

**Note:** Other OpenAI-compatible parameters (e.g., `temperature`, `top_p`, `stop`, `presence_penalty`, `frequency_penalty`) are passed through to the underlying provider.

### Streaming Response

When `stream: true`, responses are sent as Server-Sent Events:

```
data: {"id": "chatcmpl-123", "object": "chat.completion.chunk", "choices": [{"delta": {"content": "Hello"}}]}

data: {"id": "chatcmpl-123", "object": "chat.completion.chunk", "choices": [{"delta": {"content": "!"}}]}

data: [DONE]
```

### Error Responses

All errors follow OpenAI format:

```json
{
  "error": {
    "message": "Invalid API key",
    "type": "authentication_error",
    "code": 401
  }
}
```

## Supported Models

COO-LLM supports models from all configured providers:

- **OpenAI**: gpt-4o, gpt-4-turbo, gpt-3.5-turbo, etc.
- **Gemini**: gemini-1.5-pro, gemini-2.0-flash, etc.
- **Claude**: claude-3-opus, claude-3-sonnet, etc.
- **Custom**: Any provider with custom adapter

Use `provider:model` syntax for explicit selection.