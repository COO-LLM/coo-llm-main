# COO-LLM Documentation

🚀 **Intelligent Load Balancer for LLM APIs with Full OpenAI Compatibility**

COO-LLM is a high-performance reverse proxy that intelligently distributes requests across multiple LLM providers (OpenAI, Google Gemini, Anthropic Claude) and API keys. It provides seamless OpenAI API compatibility, advanced load balancing algorithms, real-time cost optimization, and enterprise-grade observability.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](https://docker.com)
[![License: DIB](https://img.shields.io/badge/License-DIB-black.svg)](https://devs-in-black.web.app/#license)
[![OpenAI Compatible](https://img.shields.io/badge/OpenAI-Compatible-green.svg)](https://platform.openai.com/docs)

## 🚀 Features

### ✨ Core Capabilities
- **🔄 Full OpenAI API Compatibility**: Drop-in replacement with identical request/response formats
- **🌐 Multi-Provider Support**: OpenAI, Google Gemini, Anthropic Claude, and custom providers
- **🧠 Intelligent Load Balancing**: Advanced algorithms (Round Robin, Least Loaded, Hybrid) with real-time optimization
- **💬 Conversation History**: Full support for multi-turn conversations and message history

### 💰 Cost & Performance Optimization
- **📊 Real-time Cost Tracking**: Monitor and optimize API costs across all providers
- **⚡ Rate Limit Management**: Sliding window rate limiting with automatic key rotation
- **📈 Performance Monitoring**: Track latency, success rates, token usage, and error patterns
- **🔄 Response Caching**: Configurable caching to reduce costs and improve performance

### 🏢 Enterprise-Ready
- **🔌 Extensible Architecture**: Plugin system for custom providers, storage backends, and logging
- **📊 Production Observability**: Prometheus metrics, structured logging, and health checks
- **⚙️ Configuration Management**: YAML-based configuration with environment variable support
- **🔒 Security**: API key masking, secure storage, and authentication controls

## 🏁 Quick Start

### Local Development

```bash
# Clone and build
git clone https://github.com/your-org/coo-llm.git
cd coo-llm
go build -o bin/coo-llm ./cmd/coo-llm

# Configure with environment variables
export OPENAI_API_KEY="sk-your-key"
export GEMINI_API_KEY="your-gemini-key"

# Run
./bin/coo-llm

# Test simple request
curl -X POST http://localhost:2906/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4o", "messages": [{"role": "user", "content": "Hello!"}]}'
```

### Docker

```bash
# Run with local build
docker run -p 2906:2906 \
  -e OPENAI_API_KEY="sk-your-key" \
  -e GEMINI_API_KEY="your-gemini-key" \
  -v $(pwd)/configs:/app/configs \
  khapu2906/coo-llm:latest
```

## 📚 Documentation

### Quick Links
- **[Introduction](docs/Intro/Overview.md)**: Overview and architecture
- **[Configuration](docs/Guides/Configuration.md)**: Complete configuration reference
- **[API Reference](docs/Reference/API.md)**: REST API documentation
- **[Load Balancing](docs/Reference/Balancer.md)**: Load balancing algorithms and policies
- **[Deployment](docs/Guides/Deployment.md)**: Installation and production deployment
- **[LangChain Demo](langchain-demo/)**: Integration examples

### Documentation Structure
- **Intro**: Overview, architecture, and getting started
- **Guides**: User guides, configuration, and deployment
- **Reference**: Technical API, configuration, and balancer reference
- **Contributing**: Development guidelines and contribution process

## 🏗️ Architecture

```
Client Applications (OpenAI SDK, LangChain, etc.)
    ↓ HTTP/JSON (OpenAI-compatible API)
COO-LLM Proxy
├── 🏺 API Layer (OpenAI-compatible REST API)
│   ├── Chat Completions (/v1/chat/completions)
│   ├── Models (/v1/models)
│   └── Admin API (/admin/v1/*)
├── ⚖️ Load Balancer (Intelligent Routing)
│   ├── Round Robin, Least Loaded, Hybrid algorithms
│   ├── Rate limiting & cost optimization
│   └── Real-time performance tracking
├── 🔌 Provider Adapters
│   ├── OpenAI (GPT-4, GPT-3.5)
│   ├── Google Gemini (1.5 Pro, etc.)
│   ├── Anthropic Claude (Opus, Sonnet)
│   └── Custom providers
├── 💾 Storage Layer
│   ├── Redis (production, with clustering)
│   ├── Memory (development)
│   ├── File-based (simple deployments)
│   └── HTTP (remote storage)
└── 📊 Observability
    ├── Structured logging (JSON)
    ├── Prometheus metrics
    ├── Response caching
    └── Health checks
    ↓
External LLM Providers (OpenAI, Gemini, Claude APIs)
```

## 📊 Key Metrics

- **🚀 Load Balancing**: Intelligent distribution across 3+ providers
- **💰 Cost Optimization**: Real-time cost tracking and automatic optimization
- **⚡ Rate Limiting**: Sliding window rate limiting with key rotation
- **📈 Performance**: Sub-millisecond routing with comprehensive monitoring
- **🔒 Security**: API key masking and secure storage
- **📊 Observability**: Prometheus metrics, structured JSON logging

---

**COO-LLM** - The Intelligent LLM API Load Balancer 🚀

*Load balance your LLM API calls across multiple providers with OpenAI compatibility, real-time cost optimization, and enterprise-grade reliability.*