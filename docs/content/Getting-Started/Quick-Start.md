---
sidebar_position: 3
tags: [getting-started, quick-start, tutorial]
description: "Get COO-LLM running in minutes with this quick start guide"
keywords: [quick start, tutorial, first steps, API, configuration]
---

# Quick Start

Welcome to COO-LLM! This guide will help you install and run COO-LLM in minutes to test the basic API.

## Prerequisites

- **Go 1.21+** (to build from source)
- **Docker** (optional, for containerized deployment)
- **API Key** from provider (OpenAI, Gemini, etc.)

## Quick Installation

### 1. Download Binary

```bash
# Download latest release
curl -L https://github.com/your-org/coo-llm/releases/latest/download/coo-llm-linux-amd64 -o coo-llm
chmod +x coo-llm
```

### 2. Create Basic Config

Create `config.yaml` file:

```yaml
version: "1.0"

server:
  listen: ":2906"
  admin_api_key: "your-admin-key"

llm_providers:
  - id: "openai"
    type: "openai"
    api_keys: ["sk-your-openai-key"]
    model: "gpt-4o"

api_keys:
  - id: "test-client"
    key: "test-key"
    allowed_providers: ["*"]
```

### 3. Run Server

```bash
export OPENAI_API_KEY="sk-your-key"
./coo-llm --config config.yaml
```

Server will run at `http://localhost:2906`.

## Test API

### Call Chat Completions

```bash
curl -X POST http://localhost:2906/api/v1/chat/completions \
  -H "Authorization: Bearer test-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "openai:gpt-4o",
    "messages": [{"role": "user", "content": "Hello, world!"}]
  }'
```

**Sample Response:**
```json
{
  "id": "chatcmpl-123",
  "object": "chat.completion",
  "created": 1677652288,
  "model": "openai:gpt-4o",
  "choices": [{
    "index": 0,
    "message": {
      "role": "assistant",
      "content": "Hello! How can I help you today?"
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

### Check Web UI

Access `http://localhost:2906/ui` in browser to view dashboard (login: admin/password).

## Basic Troubleshooting

- **"provider not found" error**: Check model format (`provider:model`)
- **"unauthorized" error**: Check API key in header
- **Server won't start**: Check if port 2906 is occupied

## Next Steps

- [Advanced configuration](../Guides/Configuration.md)
- [Production deployment](../Guides/Deployment.md)
- [Complete examples](../User-Guide/Examples.md)