## OpenAI API Compatibility Matrix

COO-LLM currently supports the most commonly used OpenAI API endpoints. Here's what's implemented and planned:

### ✅ **Implemented Endpoints**

| Endpoint | Method | Status | Description |
|----------|--------|--------|-------------|
| `/v1/chat/completions` | POST | ✅ Complete | Chat completions with streaming |
| `/v1/models` | GET | ✅ Complete | List available models |

### 🚧 **Planned Endpoints** (High Priority)

| Endpoint | Method | Priority | Description | Use Case |
|----------|--------|----------|-------------|----------|
| `/v1/embeddings` | POST | High | Text embeddings | Semantic search, RAG |
| `/v1/completions` | POST | Medium | Legacy text completions | Simple text generation |
| `/v1/images/generations` | POST | High | DALL-E image generation | AI image creation |
| `/v1/audio/transcriptions` | POST | Medium | Whisper transcription | Audio processing |
| `/v1/moderations` | POST | Low | Content moderation | Safety filtering |

### ❌ **Not Planned** (Low Priority)

| Endpoint | Reason |
|----------|--------|
| `/v1/fine-tunes` | Complex resource management |
| `/v1/files` | File storage complexity |
| `/v1/images/edits` | Limited use case |
| `/v1/images/variations` | Limited use case |
| `/v1/audio/translations` | Specialized use case |

## Implementation Priority

**Phase 1 (Current)**: Chat completions - Most used, high value
**Phase 2 (Next)**: Embeddings + Images - High demand features
**Phase 3 (Future)**: Audio + Moderation - Specialized features

## Contributing New Endpoints

To add support for new OpenAI API endpoints:

1. **Create Handler**: Add new handler in `internal/api/`
2. **Update Routes**: Register endpoint in `SetupRoutes()`
3. **Provider Support**: Implement in provider interfaces
4. **Load Balancing**: Add endpoint-specific balancing logic
5. **Testing**: Add comprehensive tests
6. **Documentation**: Update API reference

### Example: Adding Embeddings Support

```go
// internal/api/embeddings.go
type EmbeddingsHandler struct {
    selector *balancer.Selector
    reg      *provider.Registry
}

func (h *EmbeddingsHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // Parse request, select provider, make call, return response
}

// Register route
func SetupRoutes(r chi.Router, ...) {
    embeddingsHandler := NewEmbeddingsHandler(...)
    r.With(AuthMiddleware(cfg.APIKeys)).Post("/v1/embeddings", embeddingsHandler.Handle)
}
```

Would you like me to implement any of these missing endpoints?