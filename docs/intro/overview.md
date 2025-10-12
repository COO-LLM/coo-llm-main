# Overview

---
sidebar_position: 1
tags: [user-guide, getting-started]
---

TruckLLM is an intelligent reverse proxy and load balancer for Large Language Model (LLM) APIs. It provides a unified, OpenAI-compatible interface to multiple LLM providers while intelligently distributing requests across API keys and providers based on performance, cost, and rate limits.

## Key Features

### 🚀 Core Capabilities
- **OpenAI API Compatibility**: Drop-in replacement for OpenAI API with identical request/response formats
- **Multi-Provider Support**: Seamlessly route requests to OpenAI, Google Gemini, Anthropic Claude, and custom providers
- **Intelligent Load Balancing**: Advanced algorithms for optimal request distribution

### 💰 Cost & Performance Optimization
- **Real-time Cost Tracking**: Monitor and optimize API costs across providers
- **Rate Limit Management**: Automatic key rotation to avoid 429 errors
- **Performance Monitoring**: Track latency, success rates, and token usage

### 🔧 Enterprise-Ready
- **Extensible Architecture**: Plugin system for custom providers, storage, and logging
- **Production Observability**: Prometheus metrics, structured logging, and health checks
- **Configuration Management**: YAML-based configuration with hot-reload capabilities

### 📊 Advanced Features
- **Model Aliases**: Map custom model names to provider-specific models
- **Request Routing**: Smart routing based on model availability and performance
- **Admin API**: Runtime configuration and monitoring endpoints

## Use Cases

- **Cost Optimization**: Automatically choose the cheapest provider for each request
- **High Availability**: Failover between providers and keys during outages
- **Rate Limit Scaling**: Distribute load across multiple API keys
- **Multi-Cloud LLM**: Unified interface to multiple cloud LLM services
- **Development**: Easy switching between providers during development

## Architecture Overview

```
Client Apps (OpenAI SDK)
    ↓
TruckLLM Proxy
├── API Layer (OpenAI-compatible)
├── Load Balancer (Smart routing)
├── Provider Adapters (OpenAI, Gemini, Claude)
├── Storage (Redis/File/HTTP)
└── Logging (File/Prometheus/Webhook)
    ↓
External LLM Providers
```

## Quick Example

```bash
# Configure providers
cat > config.yaml << EOF
providers:
  - id: openai
    base_url: https://api.openai.com/v1
    keys:
      - secret: sk-your-key
        pricing:
          input_token_cost: 0.002
          output_token_cost: 0.01
model_aliases:
  gpt-4: openai:gpt-4
EOF

# Run TruckLLM
./truckllm -config config.yaml

# Use like OpenAI API
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your-key" \
  -d '{"model": "gpt-4", "messages": [{"role": "user", "content": "Hello"}]}'
```

## Getting Started

See [Deployment](deployment.md) for installation instructions and [Configuration](configuration.md) for setup details.