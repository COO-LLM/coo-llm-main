package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/user/coo-llm/internal/balancer"
	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/log"
	"github.com/user/coo-llm/internal/provider"
	"github.com/user/coo-llm/internal/store"
)

type ChatCompletionsHandler struct {
	selector *balancer.Selector
	logger   *log.Logger
	reg      *provider.Registry
	cfg      *config.Config
	store    store.RuntimeStore
}

func NewChatCompletionsHandler(selector *balancer.Selector, logger *log.Logger, reg *provider.Registry, cfg *config.Config, store store.RuntimeStore) *ChatCompletionsHandler {
	return &ChatCompletionsHandler{selector: selector, logger: logger, reg: reg, cfg: cfg, store: store}
}

func (h *ChatCompletionsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req map[string]any
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model, ok := req["model"].(string)
	if !ok {
		http.Error(w, "model is required", http.StatusBadRequest)
		return
	}

	// Check API key permissions
	allowedProviders, ok := r.Context().Value("allowed_providers").([]string)
	if !ok {
		http.Error(w, `{"error": {"message": "Authentication context missing", "type": "authentication_error"}}`, http.StatusInternalServerError)
		return
	}

	// Extract messages
	var messages []map[string]interface{}
	if msgs, ok := req["messages"].([]interface{}); ok && len(msgs) > 0 {
		messages = make([]map[string]interface{}, len(msgs))
		for i, msg := range msgs {
			if m, ok := msg.(map[string]interface{}); ok {
				messages[i] = m
			}
		}
	}

	// Extract prompt from last message for caching (backward compatibility)
	var prompt string
	if len(messages) > 0 {
		if content, ok := messages[len(messages)-1]["content"].(string); ok {
			prompt = content
		}
	}

	// Check cache if enabled
	if h.cfg.Policy.Cache.Enabled && prompt != "" {
		var cacheHit bool
		var cachedResp string
		var err error

		if h.cfg.Policy.Cache.SemanticEnabled {
			// Semantic caching
			cacheHit, cachedResp, err = h.checkSemanticCache(prompt)
		} else {
			// Exact match caching
			cacheKey := normalizeText(prompt)
			cachedResp, err = h.selector.GetCache(cacheKey)
			cacheHit = err == nil && cachedResp != ""
		}

		if cacheHit {
			// Return cached response
			var cached map[string]interface{}
			if json.Unmarshal([]byte(cachedResp), &cached) == nil {
				cached["cache_hit"] = true
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(cached)
				return
			}
		}
	}

	maxTokens := 1000
	if mt, ok := req["max_tokens"].(float64); ok {
		maxTokens = int(mt)
	}

	stream := false
	if s, ok := req["stream"].(bool); ok {
		stream = s
	}

	user := ""
	if u, ok := req["user"].(string); ok {
		user = u
	}

	// Retry logic
	var resp *provider.LLMResponse
	var pCfg *config.Provider
	var key *config.Key
	var modelName string
	var latency int64
	var err error

	retryCfg := h.cfg.Policy.Retry
	if retryCfg.MaxAttempts == 0 {
		retryCfg.MaxAttempts = 1 // Default no retry
	}

	for attempt := 0; attempt < retryCfg.MaxAttempts; attempt++ {
		pCfg, key, modelName, err = h.selector.SelectBest(model)
		if err != nil {
			break
		}

		// Check if provider is allowed
		allowed := false
		for _, allowedProvider := range allowedProviders {
			if allowedProvider == "*" || allowedProvider == pCfg.ID {
				allowed = true
				break
			}
		}
		if !allowed {
			err = fmt.Errorf("access denied to provider %s", pCfg.ID)
			break
		}

		prov, err := h.reg.Get(pCfg.ID)
		if err != nil {
			break
		}

		providerReq := &provider.LLMRequest{
			Prompt:    prompt,
			Messages:  messages,
			Model:     modelName,
			MaxTokens: maxTokens,
			Stream:    stream,
			User:      user,
			Params:    req,
		}

		ctx, cancel := context.WithTimeout(r.Context(), retryCfg.Timeout)
		attemptStart := time.Now()

		if stream {
			streamChan, err := prov.GenerateStream(ctx, providerReq)
			if err != nil {
				cancel()
				break
			}

			// Handle streaming response
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			flusher, ok := w.(http.Flusher)
			if !ok {
				cancel()
				http.Error(w, "Streaming not supported", http.StatusInternalServerError)
				return
			}

			go func() {
				defer cancel()
				for chunk := range streamChan {
					if chunk.Done {
						if chunk.Text != "" && !strings.HasPrefix(chunk.Text, "Error:") {
							// Send final usage data
							usageData := map[string]interface{}{
								"id":      fmt.Sprintf("%d", time.Now().UnixNano()),
								"object":  "chat.completion.chunk",
								"created": time.Now().Unix(),
								"model":   model,
								"choices": []map[string]interface{}{
									{
										"index":         0,
										"delta":         map[string]string{},
										"finish_reason": chunk.FinishReason,
									},
								},
								"usage": map[string]interface{}{
									"prompt_tokens":     0, // TODO: track properly
									"completion_tokens": 0,
									"total_tokens":      0,
								},
							}
							data, _ := json.Marshal(usageData)
							fmt.Fprintf(w, "data: %s\n\n", data)
							flusher.Flush()
						}
						fmt.Fprintf(w, "data: [DONE]\n\n")
						flusher.Flush()
						return
					}

					chunkData := map[string]interface{}{
						"id":      fmt.Sprintf("%d", time.Now().UnixNano()),
						"object":  "chat.completion.chunk",
						"created": time.Now().Unix(),
						"model":   model,
						"choices": []map[string]interface{}{
							{
								"index": 0,
								"delta": map[string]interface{}{
									"content": chunk.Text,
								},
								"finish_reason": nil,
							},
						},
					}
					data, _ := json.Marshal(chunkData)
					fmt.Fprintf(w, "data: %s\n\n", data)
					flusher.Flush()
				}
			}()

			// Update usage for streaming (simplified)
			h.selector.UpdateUsage(pCfg.ID, key.ID, "req", 1)
			return
		}

		resp, err = prov.Generate(ctx, providerReq)
		cancel()
		latency = time.Since(attemptStart).Milliseconds()

		if err == nil {
			// Extra safety check
			if resp == nil {
				err = fmt.Errorf("provider returned nil response")
			} else {
				// Success, update usage
				h.selector.UpdateUsage(pCfg.ID, key.ID, "req", 1)
				h.selector.UpdateUsage(pCfg.ID, key.ID, "input_tokens", float64(resp.InputTokens))
				h.selector.UpdateUsage(pCfg.ID, key.ID, "output_tokens", float64(resp.OutputTokens))
				h.selector.UpdateUsage(pCfg.ID, key.ID, "tokens", float64(resp.TokensUsed))
				h.selector.UpdateUsage(pCfg.ID, key.ID, "latency", float64(latency))
				break
			}
		}
		if err != nil {
			// Error, update error usage
			h.selector.UpdateUsage(pCfg.ID, key.ID, "errors", 1)
			if attempt < retryCfg.MaxAttempts-1 {
				time.Sleep(retryCfg.Interval)
			}
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate cost
	cost := float64(resp.InputTokens)*key.Pricing.InputTokenCost + float64(resp.OutputTokens)*key.Pricing.OutputTokenCost

	// Get client API key from context
	var clientKey string
	if key, ok := r.Context().Value("api_key").(string); ok {
		clientKey = key
	}

	// Store metrics for historical data
	tags := map[string]string{"provider": pCfg.ID, "key": key.ID, "model": modelName, "client_key": clientKey}
	h.store.StoreMetric("latency", float64(latency), tags, time.Now().Unix())
	h.store.StoreMetric("tokens", float64(resp.TokensUsed), tags, time.Now().Unix())
	h.store.StoreMetric("cost", cost, tags, time.Now().Unix())

	// Log the request
	reqID := fmt.Sprintf("%d", time.Now().UnixNano())
	h.logger.LogRequest(r.Context(), &log.LogEntry{
		Provider:  pCfg.ID,
		Model:     model,
		ReqID:     reqID,
		LatencyMS: latency,
		Status:    200,
		Tokens:    resp.TokensUsed,
		Cost:      cost,
		Error:     "",
	})

	// Prepare response
	openaiResp := map[string]interface{}{
		"id":      reqID,
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]string{
					"role":    "assistant",
					"content": resp.Text,
				},
				"finish_reason": resp.FinishReason,
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     resp.InputTokens,
			"completion_tokens": resp.OutputTokens,
			"total_tokens":      resp.TokensUsed,
			"cost":              cost,
		},
	}

	// Cache response if enabled
	if h.cfg.Policy.Cache.Enabled && prompt != "" {
		respJSON, _ := json.Marshal(openaiResp)
		if h.cfg.Policy.Cache.SemanticEnabled {
			h.setSemanticCache(prompt, string(respJSON))
		} else {
			cacheKey := normalizeText(prompt)
			h.selector.SetCache(cacheKey, string(respJSON), h.cfg.Policy.Cache.TTLSeconds)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openaiResp)
}

// normalizeText creates cache key from text (lowercase, remove spaces)
func normalizeText(text string) string {
	// Simple normalization: lowercase, remove extra spaces
	normalized := strings.ToLower(text)
	normalized = strings.Join(strings.Fields(normalized), "")
	return normalized
}

// checkSemanticCache checks for semantically similar cached responses
func (h *ChatCompletionsHandler) checkSemanticCache(prompt string) (bool, string, error) {
	// TODO: Implement semantic similarity search
	// For now, fall back to exact match
	cacheKey := normalizeText(prompt)
	cached, err := h.selector.GetCache(cacheKey)
	return err == nil && cached != "", cached, err
}

// setSemanticCache stores response with semantic embedding
func (h *ChatCompletionsHandler) setSemanticCache(prompt, response string) {
	// TODO: Generate embedding and store with similarity search capability
	// For now, fall back to exact match
	cacheKey := normalizeText(prompt)
	h.selector.SetCache(cacheKey, response, h.cfg.Policy.Cache.TTLSeconds)
}

// GetProviderFromModel determines the provider ID from a model name
func (h *ChatCompletionsHandler) GetProviderFromModel(model string) string {
	// 1. Check if model is in provider:model format
	if colonIndex := strings.Index(model, ":"); colonIndex != -1 {
		providerID := model[:colonIndex]
		// Check if provider ID exists
		for _, provider := range h.cfg.LLMProviders {
			if provider.ID == providerID {
				return provider.ID
			}
		}
	}

	// 2. Check model aliases
	if alias, exists := h.cfg.ModelAliases[model]; exists {
		// Parse provider from alias (format: "provider:model")
		if colonIndex := strings.Index(alias, ":"); colonIndex != -1 {
			providerTypeOrID := alias[:colonIndex]

			// Check if it's already a provider ID (from llm_providers)
			for _, provider := range h.cfg.LLMProviders {
				if provider.ID == providerTypeOrID {
					return provider.ID
				}
			}

			// If not found as ID, it might be a legacy type name
			// Try to find provider with matching type
			for _, provider := range h.cfg.LLMProviders {
				if provider.Type == providerTypeOrID {
					return provider.ID
				}
			}
		}
	}

	// 3. Fallback: try to infer from model name and find matching provider
	if strings.Contains(model, "gpt") || strings.Contains(model, "openai") {
		for _, provider := range h.cfg.LLMProviders {
			if provider.Type == "openai" {
				return provider.ID
			}
		}
	}
	if strings.Contains(model, "gemini") {
		for _, provider := range h.cfg.LLMProviders {
			if provider.Type == "gemini" {
				return provider.ID
			}
		}
	}
	if strings.Contains(model, "claude") {
		for _, provider := range h.cfg.LLMProviders {
			if provider.Type == "claude" {
				return provider.ID
			}
		}
	}

	return "" // No provider found
}

func SetupRoutes(r chi.Router, selector *balancer.Selector, logger *log.Logger, reg *provider.Registry, cfg *config.Config, store store.RuntimeStore) {
	handler := NewChatCompletionsHandler(selector, logger, reg, cfg, store)
	r.With(AuthMiddleware(cfg.APIKeys)).Post("/v1/chat/completions", handler.Handle)

	SetupModelsRoute(r, cfg)
}
