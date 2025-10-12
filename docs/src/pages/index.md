---
sidebar_position: 1
---

# Welcome to COO-LLM

Intelligent Load Balancer for LLM APIs with Full OpenAI Compatibility

## 🚀 Quick Start

Get started with COO-LLM in minutes:

```bash
# Clone and run
git clone https://github.com/your-org/coo-llm.git
cd coo-llm
go build -o bin/coo-llm cmd/coo-llm/main.go
./bin/coo-llm
```

## 📚 Documentation

- **[Introduction](./Intro/Overview.md)** - Learn about COO-LLM
- **[Configuration](./Guides/Configuration.md)** - Setup and configuration
- **[API Reference](./Reference/API.md)** - Complete API documentation
- **[Contributing](./Contributing/Guidelines.md)** - How to contribute

## ✨ Features

- **🔄 Full OpenAI API Compatibility**: Drop-in replacement
- **🌐 Multi-Provider Support**: OpenAI, Gemini, Claude, and custom providers
- **🧠 Intelligent Load Balancing**: Smart distribution based on performance
- **💬 Conversation History**: Full support for multi-turn conversations
- **📊 Real-time Monitoring**: Usage tracking and performance metrics
- **🔒 Security**: API key authentication and permissions

## 🏗️ Architecture

COO-LLM follows a modular architecture:

```
Client Apps → API Layer → Load Balancer → LLM Providers
```

Learn more in our [Architecture Guide](./Intro/Architecture.md).