---
sidebar_position: 1
---

# Introduction

Welcome to COO-LLM! This section provides an overview of what COO-LLM is and how it works.

## What is COO-LLM?

COO-LLM is an intelligent reverse proxy for Large Language Model (LLM) APIs. It provides:

- **OpenAI API Compatibility**: Drop-in replacement for OpenAI API
- **Multi-Provider Support**: Route to OpenAI, Gemini, Claude, and custom providers
- **Intelligent Load Balancing**: Smart distribution based on performance, cost, and limits
- **Enterprise Features**: Monitoring, logging, configuration management

## Quick Start

Get started in minutes:

```bash
# Clone and run
git clone https://github.com/coo-llm/coo-llm-main.git
cd coo-llm
make build && make run
```

## Key Concepts

### Providers
External LLM services like OpenAI, Google Gemini, Anthropic Claude.

### Keys
API keys for accessing providers, with rate limits and pricing.

### Models
AI models available through providers, mapped via aliases.

### Load Balancing
Automatic distribution of requests across providers and keys.

## Architecture Overview

```
Client Apps → COO-LLM → LLM Providers
    ↑              ↓
   SDKs        Load Balancer
               Metrics & Logs
```

## Next Steps

- [Overview](Intro/Overview.md) - Detailed introduction
- [Architecture](Intro/Architecture.md) - System design
- [Quick Start](./Guides/Deployment.md) - Get running