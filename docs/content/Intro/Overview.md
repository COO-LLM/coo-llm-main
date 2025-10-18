---
title: "COO-LLM Overview"
sidebar_position: 1
tags: [user-guide, getting-started, introduction, features, architecture]
description: "Learn about COO-LLM, an intelligent load balancer and reverse proxy for Large Language Model APIs with OpenAI compatibility"
keywords: [COO-LLM, LLM, load balancer, reverse proxy, OpenAI, API, AI, machine learning]
---

COO-LLM is an intelligent reverse proxy and load balancer for Large Language Model (LLM) APIs. It provides a unified, OpenAI-compatible interface to multiple LLM providers while intelligently distributing requests across API keys and providers based on performance, cost, and rate limits.

**What makes COO-LLM special?** Unlike simple API gateways, COO-LLM uses advanced algorithms to automatically choose the best provider for each request, handle rate limits gracefully, and provide deep observability into your LLM usage.

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
- **Multi-Tenant Client Management**: Dynamic API client registration with provider restrictions
- **Advanced Security**: Rate limiting, audit logging, and access control
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
COO-LLM Proxy
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
version: "1.0"
server:
  listen: ":2906"

llm_providers:
  - id: "openai-prod"
    type: "openai"
    api_keys: ["sk-your-key"]
    base_url: "https://api.openai.com"
    model: "gpt-4o"
    pricing:
      input_token_cost: 0.002
      output_token_cost: 0.01
    limits:
      req_per_min: 200
      tokens_per_min: 100000

# Use "openai-prod:gpt-4o" directly (model_aliases removed)
EOF

# Run COO-LLM
./coo-llm -config config.yaml

# Use like OpenAI API
curl -X POST http://localhost:2906/v1/chat/completions \
  -H "Authorization: Bearer your-key" \
  -d '{"model": "openai:gpt-4o", "messages": [{"role": "user", "content": "Hello"}]}'
```

## How to Use This Documentation

This documentation is organized into sections based on your role and needs:

### 📚 Documentation Structure

- **🚀 Getting Started**: Quick setup and basic concepts
- **👤 User Guide**: API usage, examples, and best practices
- **⚙️ Administrator Guide**: Configuration, monitoring, and troubleshooting
- **🔧 Developer Guide**: Architecture, API reference, and contributing
- **📖 Reference**: Complete schemas, error codes, and glossary

### 🔍 Search & Navigation

**Search Box**: Use the search box in the top navigation to find specific topics, functions, or error messages.

**Keyboard Shortcuts**:
- `Ctrl+K` (Linux/Windows) or `Cmd+K` (Mac): Open search
- `/`: Focus search box
- `↑/↓`: Navigate results
- `Enter`: Open selected result

**Tips for Effective Search**:
- Search for error messages: `"rate limit exceeded"`
- Find configuration options: `"req_per_min"`
- API endpoints: `"chat/completions"`
- Provider setup: `"openai configuration"`

**Browse by Category**:
- New users: Start with Getting Started → Quick Start
- API integration: User Guide → API Usage → Examples
- Production setup: Administrator Guide → Configuration → Monitoring
- Code contributions: Developer Guide → Contributing

### 📖 Reading Tips

- **Follow the sidebar**: Topics are ordered logically
- **Use cross-references**: Links connect related concepts
- **Check examples**: Code samples in multiple languages
- **Reference for details**: Use Reference section for complete specs

### 🆘 Getting Help

- **GitHub Issues**: Report bugs or request features
- **Discussions**: Ask questions and share experiences
- **Contributing**: Help improve the documentation

## Getting Started

See [Deployment](../Guides/Deployment.md) for installation instructions and [Configuration](../Guides/Configuration.md) for setup details.