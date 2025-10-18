---
sidebar_position: 2
tags: [user-guide, practical, implementation, tips]
description: "Practical guide to implementing COO-LLM in real applications with tips and best practices"
keywords: [practical usage, implementation, tips, tricks, patterns]
---

# Practical Usage Guide

This guide focuses on real-world implementation patterns, tips & tricks, and how to get the most out of COO-LLM in your applications.

## 🎯 Choosing the Right Model

### Use Case Decision Tree

```
Need high-quality reasoning?
├── Yes → Use GPT-4o or Claude 3 Opus
│   └── Budget constraints? → GPT-4o (cheaper)
│       └── Need multimodal? → GPT-4o Vision
└── No, need speed/cost optimization?
    ├── Creative writing/code → Gemini 1.5 Pro
    ├── Simple Q&A → Gemini 1.0 Pro or GPT-3.5
    └── Factual tasks → Any model with low temp
```

### Cost Optimization Strategies

**1. Model Fallback Chain**
```python
# Try expensive model first, fallback to cheaper ones
models_to_try = [
    "openai:gpt-4o",           # $0.03/1k tokens
    "claude:claude-3-sonnet",  # $0.015/1k tokens
    "gemini:gemini-1.5-pro"    # $0.007/1k tokens
]

for model in models_to_try:
    try:
        response = call_coo_llm(model, prompt)
        return response
    except Exception as e:
        logger.warning(f"{model} failed: {e}")
        continue
```

**2. Dynamic Model Selection**
```python
def select_model_by_complexity(text):
    word_count = len(text.split())

    if word_count < 100:
        return "gemini:gemini-1.0-pro"  # Fast & cheap
    elif word_count < 500:
        return "claude:claude-3-haiku"  # Good balance
    else:
        return "openai:gpt-4o"         # Complex reasoning needed
```

## 🚀 Implementation Patterns

### Streaming for Better UX

**Why streaming matters:**
- **User Experience**: See responses in real-time
- **Apparent Speed**: Feels faster even if total time is same
- **Memory Efficiency**: Process large responses without loading everything
- **Cancellation**: Users can stop generation early

**Implementation:**
```javascript
async function streamChat(model, messages) {
  const response = await fetch('/v1/chat/completions', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${API_KEY}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      model: model,
      messages: messages,
      stream: true,
      max_tokens: 1000
    })
  });

  const reader = response.body.getReader();
  const decoder = new TextDecoder();

  let fullContent = '';
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;

    const chunk = decoder.decode(value);
    const lines = chunk.split('\n');

    for (const line of lines) {
      if (line.startsWith('data: ')) {
        const data = line.slice(6);
        if (data === '[DONE]') return fullContent;

        try {
          const parsed = JSON.parse(data);
          const delta = parsed.choices[0]?.delta?.content || '';
          fullContent += delta;

          // Update UI in real-time
          updateUI(fullContent);
        } catch (e) {
          // Handle parse errors
        }
      }
    }
  }
}
```

### Error Handling & Resilience

**Circuit Breaker Pattern:**
```python
class COOLLMClient:
    def __init__(self):
        self.failure_count = 0
        self.last_failure_time = 0
        self.circuit_open = False

    def call_with_circuit_breaker(self, model, messages):
        if self.circuit_open:
            if time.time() - self.last_failure_time > 60:  # Reset after 1min
                self.circuit_open = False
                self.failure_count = 0
            else:
                raise Exception("Circuit breaker open")

        try:
            response = self._call_api(model, messages)
            self.failure_count = 0  # Reset on success
            return response
        except Exception as e:
            self.failure_count += 1
            self.last_failure_time = time.time()

            if self.failure_count >= 5:  # Open circuit after 5 failures
                self.circuit_open = True

            raise e
```

**Exponential Backoff with Jitter:**
```python
import random
import time

def call_with_backoff(model, messages, max_retries=5):
    for attempt in range(max_retries):
        try:
            return call_coo_llm(model, messages)
        except RateLimitError:
            if attempt == max_retries - 1:
                raise

            # Exponential backoff with jitter
            base_delay = 2 ** attempt
            jitter = random.uniform(0, 1)
            delay = base_delay + jitter

            print(f"Rate limited, waiting {delay:.2f}s...")
            time.sleep(delay)
        except Exception as e:
            # Don't retry on auth/other errors
            if "unauthorized" in str(e).lower():
                raise
            if attempt < max_retries - 1:
                time.sleep(1)
                continue
            raise
```

## 🎨 Advanced Prompting Techniques

### Chain of Thought Prompting
```python
def analyze_code_with_cot(code_snippet):
    prompt = f"""
Analyze this code step by step:

{code_snippet}

Think through your analysis:
1. What does this code do?
2. Are there any bugs or issues?
3. How could it be improved?
4. What are the performance implications?

Provide your final assessment after thinking through each point.
"""

    return call_coo_llm("openai:gpt-4o", [{"role": "user", "content": prompt}])
```

### Few-Shot Learning
```python
def classify_sentiment(text):
    examples = """
Positive: "I love this product! It's amazing!" → positive
Negative: "This is terrible, complete waste of money" → negative
Neutral: "The product works as expected" → neutral

Text: "{text}"
Classification:"""

    prompt = examples.format(text=text)
    response = call_coo_llm("gemini:gemini-1.5-pro",
                          [{"role": "user", "content": prompt}])

    return response.choices[0].message.content.strip().lower()
```

## ⚡ Performance Optimization

### Parameter Tuning Guide

| Parameter | Low Values | High Values | Use Case |
|-----------|------------|-------------|----------|
| `temperature` | 0.0-0.3 | 0.7-1.0 | Factual → Creative |
| `top_p` | 0.1-0.5 | 0.9-1.0 | Focused → Diverse |
| `max_tokens` | 50-200 | 1000-4000 | Concise → Detailed |
| `frequency_penalty` | 0.0-0.5 | 1.0-2.0 | Repetitive → Varied |

### Batch Processing
```python
async def process_batch(prompts, batch_size=5):
    """Process multiple prompts efficiently"""
    semaphore = asyncio.Semaphore(batch_size)  # Limit concurrent requests

    async def process_single(prompt):
        async with semaphore:
            return await call_coo_llm_async("gemini:gemini-1.0-pro", prompt)

    tasks = [process_single(prompt) for prompt in prompts]
    return await asyncio.gather(*tasks)
```

### Caching Strategies
```python
import hashlib

class ResponseCache:
    def __init__(self):
        self.cache = {}

    def get_cache_key(self, model, messages, params):
        # Create deterministic key from inputs
        key_data = f"{model}:{messages}:{params}"
        return hashlib.md5(key_data.encode()).hexdigest()

    def get_cached_response(self, model, messages, params):
        key = self.get_cache_key(model, messages, params)
        return self.cache.get(key)

    def cache_response(self, model, messages, params, response):
        key = self.get_cache_key(model, messages, params)
        self.cache[key] = response
```

## 🔒 Security Best Practices

### API Key Management
```python
# Rotate keys automatically
class KeyManager:
    def __init__(self):
        self.keys = {
            'openai': ['sk-key1', 'sk-key2', 'sk-key3'],
            'gemini': ['gemini-key1']
        }
        self.current_key_index = {}

    def get_next_key(self, provider):
        if provider not in self.keys:
            raise ValueError(f"No keys for provider {provider}")

        keys = self.keys[provider]
        current = self.current_key_index.get(provider, 0)
        key = keys[current]

        # Round-robin to next key
        self.current_key_index[provider] = (current + 1) % len(keys)

        return key
```

### Input Validation & Sanitization
```python
def sanitize_input(text, max_length=10000):
    """Clean and validate input text"""
    if not text or not isinstance(text, str):
        raise ValueError("Invalid input: must be non-empty string")

    # Remove potentially harmful content
    text = text.strip()
    text = re.sub(r'[\\x00-\\x1f\\x7f-\\x9f]', '', text)  # Remove control chars

    if len(text) > max_length:
        raise ValueError(f"Input too long: {len(text)} > {max_length}")

    return text
```

## 📊 Monitoring & Observability

### Usage Tracking
```python
class UsageTracker:
    def __init__(self):
        self.usage = defaultdict(lambda: {'tokens': 0, 'cost': 0.0, 'calls': 0})

    def track_usage(self, model, response):
        provider = model.split(':')[0]
        usage = response.get('usage', {})

        self.usage[provider]['tokens'] += usage.get('total_tokens', 0)
        self.usage[provider]['calls'] += 1

        # Estimate cost (simplified)
        cost_per_token = 0.002  # Example rate
        self.usage[provider]['cost'] += usage.get('total_tokens', 0) * cost_per_token

    def get_report(self):
        return dict(self.usage)
```

### Health Checks
```python
async def health_check():
    """Comprehensive health check"""
    checks = {
        'api_reachable': False,
        'providers_healthy': {},
        'response_time': None
    }

    start_time = time.time()

    try:
        # Test basic API call
        response = await call_coo_llm("gemini:gemini-1.0-pro",
                                    [{"role": "user", "content": "test"}])
        checks['api_reachable'] = True
        checks['response_time'] = time.time() - start_time

        # Test each provider
        providers = ['openai', 'gemini', 'claude']
        for provider in providers:
            try:
                await call_coo_llm(f"{provider}:test-model",
                                 [{"role": "user", "content": "test"}])
                checks['providers_healthy'][provider] = True
            except:
                checks['providers_healthy'][provider] = False

    except Exception as e:
        checks['error'] = str(e)

    return checks
```

## 🛠️ How COO-LLM Works Internally

### Load Balancing Algorithm

COO-LLM uses a **hybrid balancing strategy**:

1. **Cost-First**: Routes to cheapest available provider
2. **Performance-Aware**: Considers response times and success rates
3. **Rate Limit Conscious**: Avoids keys nearing limits
4. **Failover Ready**: Automatically switches on failures

**Decision Flow:**
```
New request arrives
├── Check cache (semantic similarity)
├── Select provider based on:
│   ├── Model availability
│   ├── Cost optimization
│   ├── Current load & rate limits
│   └── Recent performance metrics
├── Choose API key (round-robin/load-based)
├── Make request with retry logic
├── Cache successful response
└── Update metrics
```

### Rate Limiting Implementation

**Multi-Level Rate Limiting:**
- **Provider Level**: Respects external API limits
- **Key Level**: Distributes load across your keys
- **Client Level**: Per-user/application limits
- **Global Level**: Overall system protection

**Token Bucket Algorithm:**
```python
# Simplified token bucket implementation
class TokenBucket:
    def __init__(self, capacity, refill_rate):
        self.capacity = capacity
        self.tokens = capacity
        self.refill_rate = refill_rate  # tokens per second
        self.last_refill = time.time()

    def consume(self, tokens_needed):
        now = time.time()
        elapsed = now - self.last_refill

        # Refill tokens
        self.tokens = min(self.capacity,
                         self.tokens + elapsed * self.refill_rate)
        self.last_refill = now

        if self.tokens >= tokens_needed:
            self.tokens -= tokens_needed
            return True
        return False
```

This internal knowledge helps you understand why certain behaviors occur and how to optimize your usage patterns.