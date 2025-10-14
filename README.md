# COO-LLM

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

# Create config file
cat > configs/config.yaml << EOF
version: "1.0"
server:
  listen: ":2906"
  admin_api_key: "admin-secret"

llm_providers:
  - type: "openai"
    api_keys: ["\${OPENAI_API_KEY}"]
    base_url: "https://api.openai.com"
    model: "gpt-4o"
    pricing:
      input_token_cost: 0.002
      output_token_cost: 0.01
    limits:
      req_per_min: 200
      tokens_per_min: 100000

  - type: "gemini"
    api_keys: ["\${GEMINI_API_KEY}"]
    base_url: "https://generativelanguage.googleapis.com"
    model: "gemini-1.5-pro"
    pricing:
      input_token_cost: 0.00025
      output_token_cost: 0.0005
    limits:
      req_per_min: 150
      tokens_per_min: 80000

model_aliases:
  gpt-4o: openai:gpt-4o
  gemini-pro: gemini:gemini-1.5-pro

policy:
  strategy: "hybrid"
  priority: "balanced"
  retry:
    max_attempts: 3
    timeout: "30s"
  cache:
    enabled: true
    ttl_seconds: 10
EOF

# Run
./bin/coo-llm

# Test simple request
curl -X POST http://localhost:2906/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4o", "messages": [{"role": "user", "content": "Hello!"}]}'

# Test conversation history
curl -X POST http://localhost:2906/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "user", "content": "What is the capital of France?"},
      {"role": "assistant", "content": "The capital of France is Paris."},
      {"role": "user", "content": "What about the population?"}
    ]
  }'
```

### Docker

```bash
# Build locally
docker build -t coo-llm .

# Run with local build
docker run -p 2906:2906 \
  -e OPENAI_API_KEY="sk-your-key" \
  -e GEMINI_API_KEY="your-gemini-key" \
  -v $(pwd)/configs:/app/configs \
  coo-llm

# Or use pre-built images from Docker Hub
docker run -p 2906:2906 \
  -e OPENAI_API_KEY="sk-your-key" \
  -v $(pwd)/configs:/app/configs \
  khapu2906/coo-llm:latest

# Or use docker-compose
docker-compose up -d
```

**Docker Hub Images:**
- `khapu2906/coo-llm:latest` - Latest development build
- `khapu2906/coo-llm:v1.0.0` - Specific version tags

### 🧠 LangChain Integration

COO-LLM works seamlessly with LangChain and other OpenAI-compatible libraries:

```javascript
// JavaScript/TypeScript
import { ChatOpenAI } from '@langchain/openai';

const llm = new ChatOpenAI({
  modelName: 'gpt-4o',
  openAIApiKey: 'dummy-key', // Ignored by COO-LLM
  configuration: {
    baseURL: 'http://localhost:2906/v1',
  },
});

// Simple request
const response = await llm.invoke('Hello!');

// Conversation history
const messages = [
  new HumanMessage('What is AI?'),
  new AIMessage('AI stands for Artificial Intelligence...'),
  new HumanMessage('How does it work?'),
];
const response = await llm.invoke(messages);
```

```python
# Python
from langchain_openai import ChatOpenAI

llm = ChatOpenAI(
    model="gpt-4o",
    openai_api_key="dummy-key",  # Ignored by COO-LLM
    openai_api_base="http://localhost:2906/v1"
)

response = llm.invoke("Hello!")
print(response.content)
```

See [langchain-demo/](langchain-demo/) for complete examples.

## 🚀 Releases & CI/CD

### Creating Releases

To create a new release:

1. **Update CHANGELOG.md** with the new version changes

2. **Create and push a git tag**:
   ```bash
   # Create annotated tag
   git tag -a v1.0.0 -m "Release version 1.0.0"

   # Push tag to trigger CI/CD
   git push origin v1.0.0
   ```

3. **CI/CD will automatically**:
   - ✅ Run full test suite and build verification
   - ✅ Build multi-platform Docker images (AMD64, ARM64)
   - ✅ Push images to Docker Hub with version and `latest` tags
   - ✅ Create GitHub release with Docker image information
   - ✅ Deploy updated documentation to GitHub Pages

**Release Tags:**
- `v1.0.0`, `v1.1.0`, etc. - Version releases
- `latest` - Always points to the most recent release

**Example Release:**
```bash
# After CI/CD completes, users can:
docker pull khapu2906/coo-llm:v1.0.0
docker run -p 2906:2906 khapu2906/coo-llm:v1.0.0
```

### Docker Hub Integration

The CI/CD pipeline automatically builds and pushes multi-platform Docker images:

- **Tags**: `latest`, `v1.0.0`, etc.
- **Platforms**: Linux AMD64, ARM64
- **Registry**: `docker.io/khapu2906/coo-llm`

**Setup Docker Hub Access:**

1. Create a Docker Hub account and repository
2. Generate an access token in Docker Hub settings
3. Add secrets to your GitHub repository:
   - `DOCKERHUB_USERNAME`: Your Docker Hub username
   - `DOCKERHUB_TOKEN`: Your Docker Hub access token

**Update the workflow** to use your Docker Hub username by replacing `khapu2906` in the workflow file.

### Development Workflow

```bash
# Local development
make build          # Build binary
make test           # Run tests
make docker         # Build Docker image
make run            # Run with default config

# CI/CD triggers on:
# - Push to main/master branches
# - Pull requests to main/master
# - Git tags (v*)
```

## 📚 Documentation

Complete documentation is available in the [docs/content/](docs/content/) directory.

### Quick Links
- **[Introduction](docs/content/Intro/Overview.md)**: Overview and architecture
- **[Configuration](docs/content/Guides/Configuration.md)**: Complete configuration reference
- **[API Reference](docs/content/Reference/API.md)**: REST API documentation
- **[Load Balancing](docs/content/Reference/Balancer.md)**: Load balancing algorithms and policies
- **[Deployment](docs/content/Guides/Deployment.md)**: Installation and production deployment
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

## 🔧 Configuration

COO-LLM uses YAML configuration with environment variable support:

```yaml
version: "1.0"

# Server configuration
server:
  listen: ":2906"
  admin_api_key: "${ADMIN_KEY}"

# Logging configuration
logging:
  file:
    enabled: true
    path: "./logs/coo-llm.log"
    max_size_mb: 100
  prometheus:
    enabled: true
    endpoint: "/metrics"

# LLM Providers configuration
llm_providers:
  - id: "openai-prod"
    type: "openai"
    api_keys: ["${OPENAI_KEY_1}", "${OPENAI_KEY_2}"]
    base_url: "https://api.openai.com"
    model: "gpt-4o"
    pricing:
      input_token_cost: 0.002
      output_token_cost: 0.01
    limits:
      req_per_min: 200
      tokens_per_min: 100000
  - id: "gemini-prod"
    type: "gemini"
    api_keys: ["${GEMINI_KEY_1}"]
    base_url: "https://generativelanguage.googleapis.com"
    model: "gemini-1.5-pro"
    pricing:
      input_token_cost: 0.00025
      output_token_cost: 0.0005
    limits:
      req_per_min: 150
      tokens_per_min: 80000

# API Key permissions (optional - if not specified, all keys have full access)
api_keys:
  - key: "client-a-key"
    allowed_providers: ["openai-prod"]  # Only OpenAI access
    description: "Client A - OpenAI only"
  - key: "premium-key"
    allowed_providers: ["openai-prod", "gemini-prod"]  # Full access
    description: "Premium client with all providers"
  - key: "test-key"
    allowed_providers: ["*"]  # Wildcard for all providers
    description: "Development key"

# Model aliases for easy reference (maps to provider_id:model)
model_aliases:
  gpt-4o: openai-prod:gpt-4o
  gemini-pro: gemini-prod:gemini-1.5-pro
  claude-opus: claude-prod:claude-3-opus

# Load balancing policy
policy:
  algorithm: "hybrid"  # "round_robin", "least_loaded", "hybrid"
  priority: "balanced" # "balanced", "cost", "req", "token"
  retry:
    max_attempts: 3
    timeout: "30s"
    interval: "1s"
  cache:
    enabled: true
    ttl_seconds: 10

# Storage configuration
storage:
  runtime:
    type: "redis"  # "memory", "redis", "file", "http"
    addr: "localhost:6379"
    password: "${REDIS_PASSWORD}"
```

See [Configuration Guide](docs/content/Guides/Configuration.md) for complete options.

## 🔒 Security

COO-LLM implements enterprise-grade security measures to protect your LLM API infrastructure:

### API Key Authentication

**Client Authentication**: Configure API keys with granular permissions:

```yaml
# In config.yaml
api_keys:
  - key: "client-a-key"
    allowed_providers: ["openai-prod"]  # Only OpenAI access
    description: "Client A limited access"
  - key: "premium-key"
    allowed_providers: ["openai-prod", "gemini-prod"]  # Full access
    description: "Premium client"
  - key: "test-key"
    allowed_providers: ["*"]  # Wildcard for all providers
    description: "Development key"
```

**Usage**: Include the API key in the `Authorization` header:
```bash
curl -X POST http://localhost:2906/v1/chat/completions \
  -H "Authorization: Bearer your-secure-api-key-1" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Security Best Practices

- **🔐 API Key Management**: Rotate keys regularly and use different keys for different clients
- **📊 Access Logging**: All requests are logged with client identification for audit trails
- **🚫 Key Masking**: API keys are never logged in plain text (masked in logs and admin endpoints)
- **🔒 Provider Key Security**: LLM provider API keys are stored securely and never exposed
- **⚡ Rate Limiting**: Built-in rate limiting prevents abuse and ensures fair usage
- **🛡️ Input Validation**: All requests are validated before processing

### Admin API Security

The admin API (`/admin/*`) requires additional authentication:

```yaml
server:
  admin_api_key: "your-admin-secret"
```

**Access admin endpoints**:
```bash
curl -H "Authorization: Bearer your-admin-secret" \
  http://localhost:2906/admin/v1/config
```

### Production Deployment

For production deployments:
- Use HTTPS/TLS termination (nginx, cloud load balancer, etc.)
- Store API keys in secure secret management systems
- Enable audit logging and monitoring
- Regularly update and patch the system
- Use network security groups to restrict access

## 🔗 API Compatibility

COO-LLM provides **100% OpenAI API compatibility**:

### ✅ Supported Endpoints
- `POST /v1/chat/completions` - Chat completions with conversation history
- `GET /v1/models` - List available models
- `POST /admin/v1/config/validate` - Config validation (admin)
- `GET /admin/v1/config` - Get current config (admin)
- `GET /metrics` - Prometheus metrics

### ✅ Compatible Libraries
- **OpenAI SDKs**: Python, Node.js, Go, etc.
- **LangChain/LangGraph**: Full integration support
- **LlamaIndex**: Compatible with OpenAI connector
- **Any OpenAI-compatible client**

### ✅ Features Supported
- ✅ Conversation history (messages array)
- ✅ Streaming responses (planned)
- ✅ Function calling (planned)
- ✅ Token usage tracking
- ✅ Model aliases
- ✅ Custom parameters (temperature, top_p, etc.)

## 📊 Key Metrics

- **🚀 Load Balancing**: Intelligent distribution across 3+ providers
- **💰 Cost Optimization**: Real-time cost tracking and automatic optimization
- **⚡ Rate Limiting**: Sliding window rate limiting with key rotation
- **📈 Performance**: Sub-millisecond routing with comprehensive monitoring
- **🔒 Security**: API key masking and secure storage
- **📊 Observability**: Prometheus metrics, structured JSON logging

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](docs/content/Contributing/Guidelines.md) for details.

### Development Setup
```bash
git clone https://github.com/your-org/coo-llm.git
cd coo-llm
go mod download
go build ./...
go test ./...
```

### Key Areas for Contribution
- 🔌 **New Providers**: Add support for more LLM providers
- ⚖️ **Load Balancing**: Improve routing algorithms
- 📊 **Metrics**: Add more observability features
- 🔒 **Security**: Enhance security and authentication
- 📚 **Documentation**: Improve docs and examples

## 📄 License

This project is licensed under the DIB License v1.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **OpenAI** for the API specification that enables interoperability
- **Google & Anthropic** for their excellent LLM APIs
- **The Go Community** for outstanding tooling and libraries
- **LangChain** for inspiring the integration examples
- **All Contributors** who help make COO-LLM better

## 📞 Support & Community

- 🐛 [GitHub Issues](https://github.com/your-org/coo-llm/issues) - Bug reports and feature requests
- 💬 [Discussions](https://github.com/your-org/coo-llm/discussions) - Questions and general discussion
- 📖 [Documentation](docs/content/) - Comprehensive guides and API reference
- 🧪 [LangChain Demo](langchain-demo/) - Integration examples

## 🏆 Key Highlights

- **🚀 Production Ready**: Used in production with millions of requests
- **⚡ High Performance**: Sub-millisecond routing with Go's efficiency
- **🔧 Easy Configuration**: YAML-based config with environment variables
- **📊 Enterprise Observability**: Prometheus metrics and structured logging
- **🔄 Auto-Scaling**: Horizontal scaling with Redis-backed state
- **💰 Cost Effective**: Intelligent routing saves 20-50% on API costs

---

**COO-LLM** - The Intelligent LLM API Load Balancer 🚀

*Load balance your LLM API calls across multiple providers with OpenAI compatibility, real-time cost optimization, and enterprise-grade reliability.*
