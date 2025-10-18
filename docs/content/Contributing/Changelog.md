---
sidebar_position: 7
tags: [developer-guide, changelog]
---

# Changelog


All notable changes to COO-LLM will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.28] - 2025-10-18

### Added
- **CORS Support**: Full Cross-Origin Resource Sharing configuration for web applications
- **API Key IDs**: Added optional unique identifiers for API keys for better management and UI display
- **Enhanced Storage Layer**: Improved storage interfaces and provider implementations
- **Configuration Enhancements**: Extended configuration schema with new validation rules
- **Web UI Improvements**: Updated admin interface with better user experience

### Changed
- **Storage Architecture**: Refactored storage layer for better performance and reliability
- **Configuration Structure**: Enhanced config validation and environment variable handling
- **Admin Interface**: Improved web UI responsiveness and functionality

### Fixed
- **CORS Middleware**: Fixed middleware execution order and preflight request handling
- **Configuration Loading**: Resolved YAML parsing issues and validation errors
- **Storage Operations**: Fixed concurrent access issues in storage providers

### Security
- **CORS Security**: Added proper CORS headers and origin validation
- **Configuration Security**: Enhanced secure handling of sensitive configuration data

## [1.1.1] - 2025-10-17

### Added
- **CORS Support**: Cross-Origin Resource Sharing configuration for web applications
- **Conversation History Support**: Full support for multi-turn conversations with message history
- **LangChain Integration Demo**: Complete Node.js example showing LangChain compatibility
- **Sliding Window Rate Limiting**: Advanced rate limiting with Redis-backed sliding windows
- **Response Caching**: Configurable caching system to reduce costs and improve performance
- **Multiple Load Balancing Algorithms**: Round Robin, Least Loaded, and Hybrid algorithms
- **Priority-Based Routing**: Auto-configured weights based on cost, latency, requests, and tokens
- **OpenAI API Compatibility**: 100% compatible with OpenAI Chat Completions API
- **Multi-Provider Support**: OpenAI, Google Gemini, Anthropic Claude, and custom providers
- **Prometheus Metrics**: Comprehensive monitoring and alerting capabilities
- **Admin API**: Configuration management and validation endpoints
- **Environment Variable Support**: Secure configuration with `${VAR_NAME}` syntax

### Changed
- **Configuration Format**: Migrated from `providers` to `llm_providers` structure
- **Project Name**: Renamed from TruckLLM to COO-LLM
- **Response Format**: Removed custom `used_key` field for OpenAI compatibility
- **Cache TTL**: Reduced from 5 minutes to 10 seconds for better performance

### Fixed
- **Nil Pointer Panics**: Added safety checks for provider responses
- **Rate Limiting**: Fixed sliding window implementation with Redis
- **API Key Security**: Masked sensitive data in admin endpoints
- **Conversation History**: Proper handling of multi-turn conversations

### Security
- **API Key Masking**: Sensitive data is masked in logs and admin responses
- **Input Validation**: Enhanced validation for all API inputs
- **Secure Storage**: Environment variables for all secrets

## [1.0.0] - 2024-01-01

### Added
- Core load balancing engine with provider/key selection
- OpenAI-compatible REST API endpoints
- Configuration management with YAML and environment variables
- Redis-based runtime storage for metrics
- Structured logging with multiple backends
- Prometheus metrics integration
- Docker containerization
- Basic authentication and authorization
- Health check endpoints
- Request/response logging and monitoring

### Changed
- N/A (initial release)

### Deprecated
- N/A (initial release)

### Removed
- N/A (initial release)

### Fixed
- N/A (initial release)

### Security
- API key authentication
- Input validation and sanitization
- Secure logging (no sensitive data exposure)

---

## Types of Changes

- **Added** for new features
- **Changed** for changes in existing functionality
- **Deprecated** for soon-to-be removed features
- **Removed** for now removed features
- **Fixed** for any bug fixes
- **Security** in case of vulnerabilities

## Versioning

This project uses [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes

## Release Process

1. Update version numbers in code and documentation
2. Update this changelog
3. Create and push git tag
4. Build and publish release artifacts
5. Update deployment manifests
6. Announce release in relevant channels

## Contributing to Changelog

When contributing changes:

1. Add entries to the "Unreleased" section
2. Categorize changes appropriately (Added, Changed, Fixed, etc.)
3. Include issue/PR references where applicable
4. Keep descriptions concise but informative

Example:
```
### Added
- Support for custom provider plugins (#123)
- Rate limiting for API endpoints (#124)

### Fixed
- Memory leak in connection pooling (#125)
```