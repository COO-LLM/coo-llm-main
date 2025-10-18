---
sidebar_position: 3
tags: [user-guide, examples]
---

# Examples

Collection of code examples for common use cases with COO-LLM.

## Python

### Basic Chat Completion

```python
import requests

API_KEY = "your-api-key"
BASE_URL = "http://localhost:2906/v1"

def chat_completion(model, messages):
    response = requests.post(
        f"{BASE_URL}/chat/completions",
        headers={
            "Authorization": f"Bearer {API_KEY}",
            "Content-Type": "application/json"
        },
        json={
            "model": model,
            "messages": messages,
            "max_tokens": 1000,
            "temperature": 0.7
        }
    )
    return response.json()

# Usage
messages = [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "Explain quantum computing in simple terms."}
]

result = chat_completion("openai:gpt-4o", messages)
print(result["choices"][0]["message"]["content"])
```

### Streaming Response

```python
import requests
import json

def stream_chat(model, messages):
    response = requests.post(
        f"{BASE_URL}/chat/completions",
        headers={
            "Authorization": f"Bearer {API_KEY}",
            "Content-Type": "application/json"
        },
        json={
            "model": model,
            "messages": messages,
            "stream": True
        },
        stream=True
    )

    for line in response.iter_lines():
        if line:
            line = line.decode('utf-8')
            if line.startswith('data: '):
                data = line[6:]
                if data == '[DONE]':
                    break
                chunk = json.loads(data)
                content = chunk["choices"][0]["delta"].get("content", "")
                print(content, end="", flush=True)

# Usage
messages = [{"role": "user", "content": "Write a short story about AI."}]
stream_chat("gemini:gemini-1.5-pro", messages)
```

### Error Handling

```python
def safe_chat_completion(model, messages, retries=3):
    for attempt in range(retries):
        try:
            response = requests.post(
                f"{BASE_URL}/chat/completions",
                headers={
                    "Authorization": f"Bearer {API_KEY}",
                    "Content-Type": "application/json"
                },
                json={
                    "model": model,
                    "messages": messages
                },
                timeout=30
            )

            if response.status_code == 429:
                # Rate limited, wait and retry
                wait_time = 2 ** attempt  # Exponential backoff
                print(f"Rate limited, waiting {wait_time}s...")
                time.sleep(wait_time)
                continue
            elif response.status_code != 200:
                raise Exception(f"API error: {response.status_code} - {response.text}")

            return response.json()

        except requests.exceptions.RequestException as e:
            print(f"Request failed: {e}")
            if attempt < retries - 1:
                time.sleep(1)
            continue

    raise Exception("All retries failed")
```

## JavaScript/Node.js

### Basic Usage with Axios

```javascript
const axios = require('axios');

const API_KEY = 'your-api-key';
const BASE_URL = 'http://localhost:2906/v1';

async function chatCompletion(model, messages) {
  try {
    const response = await axios.post(`${BASE_URL}/chat/completions`, {
      model: model,
      messages: messages,
      max_tokens: 1000,
      temperature: 0.7
    }, {
      headers: {
        'Authorization': `Bearer ${API_KEY}`,
        'Content-Type': 'application/json'
      }
    });

    return response.data;
  } catch (error) {
    console.error('Error:', error.response?.data || error.message);
    throw error;
  }
}

// Usage
const messages = [
  { role: 'system', content: 'You are a coding assistant.' },
  { role: 'user', content: 'Write a Python function to reverse a string.' }
];

chatCompletion('openai:gpt-4o', messages)
  .then(result => {
    console.log(result.choices[0].message.content);
  });
```

### Streaming with Fetch API

```javascript
async function streamChat(model, messages) {
  const response = await fetch(`${BASE_URL}/chat/completions`, {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${API_KEY}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      model: model,
      messages: messages,
      stream: true
    })
  });

  const reader = response.body.getReader();
  const decoder = new TextDecoder();

  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    const chunk = decoder.decode(value);
    const lines = chunk.split('\n');

    for (const line of lines) {
      if (line.startsWith('data: ')) {
        const data = line.slice(6);
        if (data === '[DONE]') return;

        try {
          const parsed = JSON.parse(data);
          const content = parsed.choices[0]?.delta?.content || '';
          process.stdout.write(content);
        } catch (e) {
          // Ignore parse errors
        }
      }
    }
  }
}
```

## cURL

### Simple Request

```bash
curl -X POST http://localhost:2906/api/v1/chat/completions \
  -H "Authorization: Bearer your-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai:gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### With Custom Parameters

```bash
curl -X POST http://localhost:2906/api/v1/chat/completions \
  -H "Authorization: Bearer your-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude:claude-3-opus-20240229",
    "messages": [
      {"role": "system", "content": "Be concise."},
      {"role": "user", "content": "Explain recursion."}
    ],
    "max_tokens": 500,
    "temperature": 0.3
  }'
```

### Testing Rate Limits

```bash
# Send multiple requests quickly
for i in {1..10}; do
  curl -X POST http://localhost:2906/api/v1/chat/completions \
    -H "Authorization: Bearer your-key" \
    -H "Content-Type: application/json" \
    -d '{"model": "openai:gpt-4o", "messages": [{"role": "user", "content": "Hi"}]}' \
    -w "Status: %{http_code}\n" -o /dev/null &
done
wait
```

## Advanced Examples

### Multi-Provider Fallback

```python
def smart_completion(messages, preferred_providers=['openai', 'gemini', 'claude']):
    for provider in preferred_providers:
        try:
            model = f"{provider}:gpt-4o" if provider == 'openai' else f"{provider}:gemini-1.5-pro" if provider == 'gemini' else f"{provider}:claude-3-opus-20240229"
            return chat_completion(model, messages)
        except Exception as e:
            print(f"{provider} failed: {e}")
            continue
    raise Exception("All providers failed")
```

### Cost Tracking

```python
def completion_with_cost_tracking(model, messages):
    start_time = time.time()
    result = chat_completion(model, messages)
    end_time = time.time()

    usage = result.get('usage', {})
    prompt_tokens = usage.get('prompt_tokens', 0)
    completion_tokens = usage.get('completion_tokens', 0)

    # Estimate cost (simplified)
    cost_per_token = 0.002 / 1000  # Example for GPT-4
    estimated_cost = (prompt_tokens + completion_tokens) * cost_per_token

    return {
        'response': result,
        'cost': estimated_cost,
        'latency': end_time - start_time
    }
```