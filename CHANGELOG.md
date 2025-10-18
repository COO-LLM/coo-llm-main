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

## [1.2.0] - 2025-10-17

### Added
- **xAI Grok Provider**: Full integration with xAI's Grok models with OpenAI-compatible API
- **Streaming Responses**: Real-time streaming support for chat completions with Server-Sent Events
- **Semantic Caching**: Configurable semantic similarity-based caching with embedding support
- **Advanced Rate Limiting**: Per-user identification, burst handling (10% overage), and dynamic adjustments
- **User Field Support**: Added `user` parameter in requests for per-user tracking

### Changed
- **Provider Interface**: Added `GenerateStream` method for streaming support
- **Rate Limiting Logic**: Enhanced with burst allowance in `isRateLimited`
- **Cache Configuration**: Added semantic caching options (embedding_model, similarity_threshold)
- **Multi-Provider Support**: Now includes xAI Grok alongside OpenAI, Gemini, and Claude

### Fixed
- **Mutex Deadlock**: Fixed deadlock in `NextAPIKey` by moving `InitUsages` outside lock
- **Go Vet Issues**: Fixed mutex copying and unkeyed struct literals
- **Docusaurus Build**: Removed inline style causing React JSX error

## [1.1.1] - 2025-10-17

### Added
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

## [1.0.2] - 2025-10-12

### Added
- Initial release of COO-LLM
- Basic load balancing across OpenAI API keys
- YAML configuration support
- File-based logging
- Docker support
- Basic health checks

### Known Issues
- Limited to OpenAI provider only
- No conversation history support
- Basic round-robin load balancing
- No caching or advanced rate limiting