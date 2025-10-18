---
sidebar_position: 2
tags: [getting-started, installation, docker, deployment]
description: "Install COO-LLM using Docker with multiple deployment options for development and production"
keywords: [installation, docker, container, deployment, setup]
---

# Installation

COO-LLM is distributed as a Docker image for easy deployment. Choose the installation method that best fits your needs.

## Docker Installation

### Method 1: Quick Start (No Config)

Start COO-LLM immediately with default settings for testing:

```bash
docker run -d \
  --name coo-llm \
  -p 2906:2906 \
  -e OPENAI_API_KEY="your-openai-key" \
  khapu2906/coo-llm:latest
```

**What's included:**
- Basic OpenAI provider configuration
- SQLite database storage
- Default admin credentials (admin/password)
- Web UI enabled

Access at: `http://localhost:2906/ui`

### Method 2: With Custom Config

Mount your configuration file for production use:

```bash
# Create config directory
mkdir coo-llm-config

# Create config.yaml
cat > coo-llm-config/config.yaml << EOF
version: "1.0"
server:
  listen: ":2906"
  admin_api_key: "your-admin-key"

llm_providers:
  - id: "openai"
    type: "openai"
    api_keys: ["your-openai-key"]
    model: "gpt-4o"

api_keys:
  - id: "default-client"
    key: "your-client-key"
    allowed_providers: ["*"]
EOF

# Run with config
docker run -d \
  --name coo-llm \
  -p 2906:2906 \
  -v $(pwd)/coo-llm-config:/config \
  khapu2906/coo-llm:latest \
  --config /config/config.yaml
```

### Method 3: With Custom Web UI

Mount both config and custom web UI:

```bash
# Assuming you have custom UI build in ./my-webui/
docker run -d \
  --name coo-llm \
  -p 2906:2906 \
  -v $(pwd)/coo-llm-config:/config \
  -v $(pwd)/my-webui:/webui \
  -e COO_WEB_UI_PATH="/webui" \
  khapu2906/coo-llm:latest \
  --config /config/config.yaml
```

## Advanced Docker Usage

### With Redis Storage

For production with persistent metrics:

```bash
# Start Redis first
docker run -d --name redis -p 6379:6379 redis:alpine

# Run COO-LLM with Redis
docker run -d \
  --name coo-llm \
  --link redis \
  -p 2906:2906 \
  -v $(pwd)/coo-llm-config:/config \
  khapu2906/coo-llm:latest \
  --config /config/config.yaml
```

Config for Redis:
```yaml
storage:
  runtime:
    type: "redis"
    addr: "redis:6379"
```

### Docker Compose Example

Create `docker-compose.yml`:

```yaml
version: '3.8'
services:
  coo-llm:
    image: khapu2906/coo-llm:latest
    ports:
      - "2906:2906"
    volumes:
      - ./config:/config
      - ./webui:/webui  # Optional custom UI
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - COO_WEB_UI_PATH=/webui  # If using custom UI
    command: ["--config", "/config/config.yaml"]
    depends_on:
      - redis

  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
```

Run with: `docker-compose up -d`

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENAI_API_KEY` | OpenAI API key | - |
| `GEMINI_API_KEY` | Gemini API key | - |
| `CLAUDE_API_KEY` | Claude API key | - |
| `COO_WEB_UI_PATH` | Custom web UI path | - |
| `PORT` | Server port | 2906 |

## Verification

After installation, verify COO-LLM is running:

```bash
# Check container status
docker ps | grep coo-llm

# Test API
curl -X POST http://localhost:2906/api/v1/chat/completions \
  -H "Authorization: Bearer your-client-key" \
  -H "Content-Type: application/json" \
  -d '{"model": "openai:gpt-4o", "messages": [{"role": "user", "content": "Hello"}]}'

# Check Web UI
curl http://localhost:2906/health
```

## Next Steps

- [Configuration Guide](../Guides/Configuration.md) for detailed config options
- [Web UI Guide](../Administrator-Guide/Web-UI.md) for custom UI development
- [Quick Start](../Getting-Started/Quick-Start.md) for your first API call