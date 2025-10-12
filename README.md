# TruckLLM

An intelligent reverse proxy for Large Language Model (LLM) APIs with full OpenAI API compatibility. Load balances across multiple API keys and providers (OpenAI, Gemini, Claude, etc.) while providing flexible logging, storage, and YAML configuration like Docker Compose.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](https://docker.com)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 🚀 Features

### Core Capabilities
- **OpenAI API Compatible**: Drop-in replacement for OpenAI API with identical request/response formats
- **Multi-Provider Support**: Seamlessly route to OpenAI, Google Gemini, Anthropic Claude, and custom providers
- **Intelligent Load Balancing**: Advanced algorithms for optimal request distribution based on performance, cost, and rate limits

### Cost & Performance Optimization
- **Real-time Cost Tracking**: Monitor and optimize API costs across providers
- **Rate Limit Management**: Automatic key rotation to avoid 429 errors
- **Performance Monitoring**: Track latency, success rates, and token usage

### Enterprise-Ready
- **Extensible Architecture**: Plugin system for custom providers, storage, and logging
- **Production Observability**: Prometheus metrics, structured logging, and health checks
- **Configuration Management**: YAML-based configuration with hot-reload capabilities

## 🏁 Quick Start

### Local Development

```bash
# Clone and build
git clone https://github.com/your-org/truckllm.git
cd truckllm
go build -o bin/truckllm ./cmd/truckllm

# Configure
export OPENAI_API_KEY="sk-your-key"
cat > configs/config.yaml << EOF
version: "1.0"
server:
  listen: ":8080"
providers:
  - id: openai
    base_url: "https://api.openai.com/v1"
    keys:
      - secret: "\${OPENAI_API_KEY}"
model_aliases:
  gpt-4: openai:gpt-4
EOF

# Run
./bin/truckllm -config configs/config.yaml

# Test
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer test" \
  -d '{"model": "gpt-4", "messages": [{"role": "user", "content": "Hello"}]}'
```

### Docker

```bash
docker run -p 8080:8080 \
  -e OPENAI_API_KEY="sk-your-key" \
  -v $(pwd)/configs:/app/configs \
  truckllm:latest
```

## 📚 Documentation

Complete documentation is available at [docs/](docs/) or online at [truckllm.dev](https://truckllm.dev).

### Quick Links
- **[Getting Started](docs/intro/intro.md)**: Quick start guide
- **[Configuration](docs/guides/configuration.md)**: Complete configuration reference
- **[API Reference](docs/reference/api.md)**: REST API documentation
- **[Deployment](docs/guides/deployment.md)**: Installation and production deployment
- **[Contributing](docs/contributing/guidelines.md)**: Development guidelines

### Documentation Structure
- **Intro**: Overview and architecture
- **Guides**: User guides and tutorials
- **Reference**: Technical API and configuration reference
- **Contributing**: Development and contribution guidelines

## 🏗️ Architecture

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

## 🔧 Configuration

TruckLLM uses YAML configuration for all settings:

```yaml
version: "1.0"
server:
  listen: ":8080"
providers:
  - id: openai
    base_url: "https://api.openai.com/v1"
    keys:
      - secret: "${OPENAI_API_KEY}"
        pricing:
          input_token_cost: 0.002
          output_token_cost: 0.01
model_aliases:
  gpt-4: openai:gpt-4
policy:
  strategy: "hybrid"
  hybrid_weights:
    cost_ratio: 0.3
    latency: 0.2
    error_score: 0.2
    req_ratio: 0.2
    token_ratio: 0.1
```

See [Configuration](docs/configuration.md) for complete options.

## 📊 Key Metrics

- **Load Balancing**: Automatic distribution across API keys and providers
- **Cost Optimization**: Real-time cost tracking and optimization
- **Rate Limiting**: Intelligent key rotation to avoid limits
- **Observability**: Comprehensive metrics and structured logging

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](docs/contributing.md) for details.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- OpenAI for the API specification
- The Go community for excellent tooling
- All contributors and users

## 📞 Support

- [GitHub Issues](https://github.com/your-org/truckllm/issues) for bugs and features
- [Discussions](https://github.com/your-org/truckllm/discussions) for questions
- [Documentation](docs/) for detailed guides

---

**TruckLLM** - Intelligent LLM API Load Balancing 🚀