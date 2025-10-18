---
sidebar_position: 4
tags: [getting-started, first-steps, api-keys, usage, error-handling]
description: "Set up API keys, make your first requests, and handle common errors"
keywords: [API keys, authentication, error handling, troubleshooting, usage]
---

# First Steps

Now that you have COO-LLM running, let's set up API keys, make your first requests, and handle common errors.

## API Key Setup

COO-LLM requires API keys for authentication and provider access. There are two types of keys:

### 1. Provider API Keys

These are your actual LLM provider keys (OpenAI, Gemini, Claude). They are used to make requests to external APIs.

**Environment Variables:**
```bash
export OPENAI_API_KEY="sk-your-openai-key"
export GEMINI_API_KEY="your-gemini-key"
export CLAUDE_API_KEY="your-claude-key"
```

**Configuration File:**
```yaml
llm_providers:
  - id: "openai-prod"
    type: "openai"
    api_keys: ["sk-your-openai-key"]
    model: "gpt-4o"
  - id: "gemini"
    type: "gemini"
    api_keys: ["your-gemini-key"]
    model: "gemini-pro"
```

### 2. Client API Keys

These are keys your applications use to authenticate with COO-LLM. They control which providers can be accessed.

**Configuration:**
```yaml
api_keys:
  - id: "client-001"
    key: "client-key-1"
    allowed_providers: ["openai-prod", "gemini"]
    limits:
      req_per_min: 100
      tokens_per_min: 50000
  - key: "admin-key"
    allowed_providers: ["*"]  # All providers
    admin: true
```

## Basic Usage

### Making API Calls

Use the OpenAI-compatible API format:

```bash
# Chat completions
curl -X POST http://localhost:2906/api/v1/chat/completions \
  -H "Authorization: Bearer client-key-1" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai-prod:gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}],
    "max_tokens": 100
  }'
```

### Model Format

Models are specified as `provider_id:model_name`:

- `openai-prod:gpt-4o` - Use GPT-4o from openai-prod provider
- `gemini:gemini-pro` - Use Gemini Pro from gemini provider
- `claude:claude-3-sonnet-20240229` - Use Claude 3 Sonnet

### Response Format

COO-LLM returns standard OpenAI API responses:

```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "openai-prod:gpt-4o",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help you?"
    },
    "finish_reason": "stop"
  }],
  "usage": {
    "prompt_tokens": 13,
    "completion_tokens": 7,
    "total_tokens": 20
  }
}
```

## Error Handling

### Common Errors

**401 Unauthorized:**
```json
{
  "error": {
    "message": "invalid api key",
    "type": "authentication_error",
    "code": "invalid_api_key"
  }
}
```
**Solution:** Check your client API key in the Authorization header.

**403 Forbidden:**
```json
{
  "error": {
    "message": "provider not allowed",
    "type": "authorization_error",
    "code": "provider_not_allowed"
  }
}
```
**Solution:** Ensure the provider is in `allowed_providers` for your key.

**429 Rate Limited:**
```json
{
  "error": {
    "message": "rate limit exceeded",
    "type": "rate_limit_error",
    "code": "rate_limit_exceeded"
  }
}
```
**Solution:** COO-LLM will automatically retry with different keys. Increase limits in config if needed.

**400 Bad Request:**
```json
{
  "error": {
    "message": "model 'invalid-model' not found",
    "type": "invalid_request_error",
    "code": "model_not_found"
  }
}
```
**Solution:** Check model format (`provider:model`) and ensure provider exists.

### Error Codes Reference

| Code | Description | Solution |
|------|-------------|----------|
| `invalid_api_key` | Client key invalid | Check Authorization header |
| `provider_not_allowed` | Provider access denied | Update `allowed_providers` |
| `rate_limit_exceeded` | Too many requests | Wait or increase limits |
| `model_not_found` | Invalid model format | Use `provider:model` format |
| `provider_error` | Upstream provider error | Check provider status |
| `configuration_error` | Config issue | Check server logs |

### Handling Errors in Code

**Python Example:**
```python
import openai
import time

client = openai.OpenAI(
    api_key="client-key-1",
    base_url="http://localhost:2906/api/v1"
)

def safe_completion(messages, retries=3):
    for attempt in range(retries):
        try:
            response = client.chat.completions.create(
                model="openai-prod:gpt-4o",
                messages=messages
            )
            return response
        except openai.RateLimitError:
            if attempt < retries - 1:
                time.sleep(2 ** attempt)  # Exponential backoff
                continue
            raise
        except openai.AuthenticationError:
            raise  # Don't retry auth errors
```

**JavaScript Example:**
```javascript
async function safeCompletion(messages) {
  const maxRetries = 3;
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      const response = await fetch('http://localhost:2906/api/v1/chat/completions', {
        method: 'POST',
        headers: {
          'Authorization': 'Bearer client-key-1',
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          model: 'openai-prod:gpt-4o',
          messages: messages
        })
      });

      if (!response.ok) {
        const error = await response.json();
        if (response.status === 429 && attempt < maxRetries - 1) {
          // Rate limited, retry with backoff
          await new Promise(resolve => setTimeout(resolve, 1000 * 2 ** attempt));
          continue;
        }
        throw new Error(`${response.status}: ${error.error.message}`);
      }

      return await response.json();
    } catch (error) {
      if (attempt === maxRetries - 1) throw error;
    }
  }
}
```

## Monitoring Your Usage

### Check Health Status

```bash
curl http://localhost:2906/health
```

### View Metrics (Admin API)

```bash
curl -H "Authorization: Bearer admin-key" \
  http://localhost:2906/admin/metrics
```

### Web UI Dashboard

Access `http://localhost:2906/ui` to view:
- Request counts and success rates
- Cost tracking per provider
- Active clients and providers
- Real-time logs

## Next Steps

- [API Usage Guide](../User-Guide/API-Usage.md) - Complete API reference
- [Configuration Guide](../Guides/Configuration.md) - Advanced setup options
- [Examples](../User-Guide/Examples.md) - More code samples