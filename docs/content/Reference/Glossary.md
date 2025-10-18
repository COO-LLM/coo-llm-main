---
sidebar_position: 3
tags: [reference, glossary]
---

# Glossary

Definitions of key terms and concepts in COO-LLM.

## A

**API Key**: Authentication token for accessing COO-LLM APIs. Configured in `api_keys` section with provider permissions.

**Algorithm**: Load balancing method. Options: `round_robin`, `least_loaded`, `hybrid`.

## B

**Balancer**: Component responsible for selecting providers and keys based on policies and usage metrics.

**Base URL**: Provider's API endpoint. Defaults to official URLs but can be customized.

## C

**Caching**: Response caching to reduce API calls and improve performance. Configured in `policy.cache`.

**Client**: Application using COO-LLM API. Identified by API key.

**Config**: YAML configuration file defining providers, policies, and settings.

**Cost Estimation**: Calculated cost based on token usage and provider pricing.

## E

**E2E Tests**: End-to-end tests validating complete request flows.

## F

**Failover**: Automatic switching to alternative providers/keys when primary fails.

## H

**Health Check**: Endpoint (`/health`) for monitoring server status.

**Hybrid Scoring**: Load balancing algorithm combining multiple metrics (cost, latency, errors).

## K

**Key**: API key for a provider. Supports multiple keys per provider for load balancing.

## L

**Latency**: Request processing time, measured in milliseconds.

**Limits**: Rate limiting configuration per key (requests/minute, tokens/minute).

**Load Balancing**: Distribution of requests across multiple providers/keys.

## M

**Metrics**: Performance and usage statistics collected by COO-LLM.

**Model**: AI model identifier (e.g., `gpt-4o`, `claude-3-opus`).

**Model Alias**: Short name mapping to full `provider:model` syntax (deprecated).

## P

**Policy**: Configuration defining load balancing behavior and algorithms.

**Pricing**: Cost structure per provider (input/output token costs).

**Provider**: LLM service (OpenAI, Gemini, Claude). Configured in `llm_providers`.

## R

**Rate Limiting**: Request throttling based on configured limits.

**Request ID**: Unique identifier for each API request, used for tracing.

**Retry**: Automatic retry on failures with configurable attempts and intervals.

**Round Robin**: Simple load balancing algorithm cycling through options sequentially.

## S

**Scoring**: Algorithm calculating preference scores for provider/key selection.

**Session**: Time window for rate limiting (e.g., "1h", "1d").

**Storage**: Backend for runtime data (metrics, cache). Types: memory, Redis, InfluxDB, etc.

**Streaming**: Real-time response streaming for long outputs.

## T

**Token**: Unit of text processing. Used for pricing and rate limiting.

**TTL (Time To Live)**: Cache expiration time in seconds.

## U

**Usage Tracking**: Monitoring of requests, tokens, and costs per provider/key.

## W

**Web UI**: Built-in web interface for monitoring and configuration.

**Weights**: Scoring multipliers in hybrid algorithm (0.0-1.0 range).