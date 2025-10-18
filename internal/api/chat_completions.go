package api

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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

	// Check if the requested model/provider is allowed for this API key
	providerID := h.GetProviderFromModel(model)
	if providerID != "" {
		allowed := false
		for _, allowedProvider := range allowedProviders {
			if allowedProvider == "*" || allowedProvider == providerID {
				allowed = true
				break
			}
		}
		if !allowed {
			http.Error(w, `{"error": {"message": "Provider not allowed for this API key", "type": "authentication_error"}}`, http.StatusForbidden)
			return
		}
	}

	// Extract messages
	var messages []map[string]any
	if msgs, ok := req["messages"].([]any); ok && len(msgs) > 0 {
		messages = make([]map[string]any, len(msgs))
		for i, msg := range msgs {
			if m, ok := msg.(map[string]any); ok {
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
			var cached map[string]any
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
		pCfg, _, modelName, err = h.selector.SelectBest(model)
		if err != nil {
			break
		}
		if pCfg == nil {
			err = fmt.Errorf("no provider selected")
			break
		}

		// Check for recommended key in store
		recommendKey := ""
		cacheKey := "recommend_" + pCfg.ID
		if cached, cacheErr := h.selector.GetCache(cacheKey); cacheErr == nil && cached != "" {
			recommendKey = cached
			// Delete from cache immediately
			h.selector.SetCache(cacheKey, "", 0)
		}

		// Select key
		if recommendKey != "" {
			// Find the recommended key
			key = nil
			for i := range pCfg.Keys {
				if pCfg.Keys[i].ID == recommendKey {
					key = &pCfg.Keys[i]
					break
				}
			}
			if key == nil {
				// Recommended key not found, fall back to algorithm
				key, err = h.selector.SelectKeyForProvider(pCfg, modelName)
				if err != nil {
					break
				}
			}
		} else {
			// Use algorithm to select key
			key, err = h.selector.SelectKeyForProvider(pCfg, modelName)
			if err != nil {
				break
			}
		}

		if key == nil {
			err = fmt.Errorf("no key selected")
			break
		}

		// Update req usage immediately to avoid spam on this key
		h.selector.UpdateUsage(pCfg.ID, key.ID, "req", 1)

		// Limit max tokens by provider's limit
		limitedMaxTokens := maxTokens
		if pCfg.Limits.MaxTokens > 0 && limitedMaxTokens > pCfg.Limits.MaxTokens {
			limitedMaxTokens = pCfg.Limits.MaxTokens
		}

		prov, err := h.reg.Get(pCfg.ID)
		if err != nil {
			break
		}

		providerReq := &provider.LLMRequest{
			Prompt:    prompt,
			Messages:  messages,
			Model:     modelName,
			MaxTokens: limitedMaxTokens,
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
							usageData := map[string]any{
								"id":      fmt.Sprintf("%d", time.Now().UnixNano()),
								"object":  "chat.completion.chunk",
								"created": time.Now().Unix(),
								"model":   model,
								"choices": []map[string]any{
									{
										"index":         0,
										"delta":         map[string]string{},
										"finish_reason": chunk.FinishReason,
									},
								},
								"usage": map[string]any{
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

					chunkData := map[string]any{
						"id":      fmt.Sprintf("%d", time.Now().UnixNano()),
						"object":  "chat.completion.chunk",
						"created": time.Now().Unix(),
						"model":   model,
						"choices": []map[string]any{
							{
								"index": 0,
								"delta": map[string]any{
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

			// Update usage for streaming (req already updated when selected)

			// Calculate and cache recommended key for next time
			if recommended := h.selector.GetRecommendedKey(pCfg, modelName); recommended != nil {
				cacheKey := "recommend_" + pCfg.ID
				h.selector.SetCache(cacheKey, recommended.ID, 3600) // 1 hour TTL
			}
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
				// Success, update usage (req already updated when selected)
				if key != nil {
					h.selector.UpdateUsage(pCfg.ID, key.ID, "input_tokens", float64(resp.InputTokens))
					h.selector.UpdateUsage(pCfg.ID, key.ID, "output_tokens", float64(resp.OutputTokens))
					h.selector.UpdateUsage(pCfg.ID, key.ID, "tokens", float64(resp.TokensUsed))
					h.selector.UpdateUsage(pCfg.ID, key.ID, "latency", float64(latency))
				}

				// Calculate and cache recommended key for next time
				if recommended := h.selector.GetRecommendedKey(pCfg, modelName); recommended != nil {
					cacheKey := "recommend_" + pCfg.ID
					h.selector.SetCache(cacheKey, recommended.ID, 3600) // 1 hour TTL
				}
				break
			}
		}
		if err != nil {
			// Error, update error usage
			h.logger.LogRequest(r.Context(), &log.LogEntry{
				Provider:  pCfg.ID,
				Model:     model,
				ReqID:     fmt.Sprintf("%d", time.Now().UnixNano()),
				LatencyMS: latency,
				Status:    500,
				Tokens:    0,
				Cost:      0,
				Error:     err.Error(),
			})
			if key != nil {
				h.selector.UpdateUsage(pCfg.ID, key.ID, "errors", 1)
			}
			if attempt < retryCfg.MaxAttempts-1 {
				time.Sleep(retryCfg.Interval)
			}
		}
	}

	// If primary provider failed and fallback is enabled, try fallback providers
	if err != nil && h.cfg.Policy.Fallback.Enabled && !stream && pCfg != nil {
		fallbackProviders := h.getFallbackProviders(pCfg.ID, modelName)
		for _, fallbackID := range fallbackProviders {
			if fallbackID == pCfg.ID {
				continue // Skip same provider
			}

			// Try fallback provider
			fallbackPCfg, fallbackKey, fallbackModelName, fallbackResp, fallbackErr := h.tryFallbackProvider(fallbackID, modelName, req, stream)
			if fallbackErr == nil && fallbackResp != nil {
				// Fallback success, use this response
				pCfg = fallbackPCfg
				key = fallbackKey
				modelName = fallbackModelName
				resp = fallbackResp
				err = nil
				break
			}
		}
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resp == nil {
		http.Error(w, "Provider returned nil response", http.StatusInternalServerError)
		return
	}

	// Calculate cost (pricing is per 1 million tokens)
	var cost float64
	if key != nil {
		cost = (float64(resp.InputTokens)*pCfg.Pricing.InputTokenCost + float64(resp.OutputTokens)*pCfg.Pricing.OutputTokenCost) / 1000000
	}

	// Get client API key from context
	var clientKey string
	if key, ok := r.Context().Value("api_key").(string); ok {
		clientKey = key
	}

	// Store metrics for historical data
	providerName := pCfg.Name
	if providerName == "" {
		providerName = pCfg.ID
	}
	tags := map[string]string{"provider": providerName, "key": key.ID, "model": modelName, "client_key": clientKey}
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
	openaiResp := map[string]any{
		"id":      reqID,
		"object":  "chat.completion",
		"created": time.Now().Unix(),
		"model":   model,
		"choices": []map[string]any{
			{
				"index": 0,
				"message": map[string]string{
					"role":    "assistant",
					"content": resp.Text,
				},
				"finish_reason": resp.FinishReason,
			},
		},
		"usage": map[string]any{
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

// getFallbackProviders returns list of fallback provider IDs to try
func (h *ChatCompletionsHandler) getFallbackProviders(primaryID, modelName string) []string {
	fallbackCfg := h.cfg.Policy.Fallback

	// If specific fallback providers configured, use them
	if len(fallbackCfg.Providers) > 0 {
		maxCount := int(math.Min(float64(len(fallbackCfg.Providers)), float64(fallbackCfg.MaxProviders)))
		return fallbackCfg.Providers[:maxCount]
	}

	// Otherwise, try to find providers that might support similar models
	var candidates []string
	for _, lp := range h.cfg.LLMProviders {
		if lp.ID != primaryID {
			// Simple heuristic: prefer providers with similar model names or OpenAI-compatible ones
			if strings.Contains(lp.Model, modelName) ||
				strings.Contains(modelName, lp.Model) ||
				lp.Type == "openai" {
				candidates = append(candidates, lp.ID)
			}
		}
	}

	// Limit to MaxProviders
	if len(candidates) > fallbackCfg.MaxProviders {
		candidates = candidates[:fallbackCfg.MaxProviders]
	}

	return candidates
}

// tryFallbackProvider attempts to use a fallback provider and returns the response
func (h *ChatCompletionsHandler) tryFallbackProvider(providerID, modelName string, req map[string]any, stream bool) (*config.Provider, *config.Key, string, *provider.LLMResponse, error) {
	// Select fallback provider
	pCfg, key, resolvedModelName, err := h.selector.SelectBest(providerID + ":" + modelName)
	if err != nil {
		return nil, nil, "", nil, err
	}

	// Get provider instance
	prov, err := h.reg.Get(pCfg.ID)
	if err != nil {
		return nil, nil, "", nil, err
	}

	// Prepare request
	maxTokens := 1000
	if mt, ok := req["max_tokens"].(float64); ok {
		maxTokens = int(mt)
	}

	user := ""
	if u, ok := req["user"].(string); ok {
		user = u
	}

	providerReq := &provider.LLMRequest{
		Prompt:    "", // Will be set from messages
		Messages:  nil,
		Model:     resolvedModelName,
		MaxTokens: maxTokens,
		Stream:    stream,
		User:      user,
	}

	// Convert messages
	if msgs, ok := req["messages"].([]any); ok {
		providerReq.Messages = make([]map[string]any, len(msgs))
		for i, msg := range msgs {
			if msgMap, ok := msg.(map[string]any); ok {
				providerReq.Messages[i] = msgMap
			}
		}
	}

	// Try the request
	ctx, cancel := context.WithTimeout(context.Background(), h.cfg.Policy.Retry.Timeout)
	defer cancel()

	var resp *provider.LLMResponse
	if stream {
		// For fallback, we don't handle streaming yet - just test if provider works
		_, err = prov.GenerateStream(ctx, providerReq)
	} else {
		resp, err = prov.Generate(ctx, providerReq)
	}

	if err != nil {
		// Update error usage for fallback provider
		if key != nil {
			h.selector.UpdateUsage(pCfg.ID, key.ID, "errors", 1)
		}
		return nil, nil, "", nil, err
	}

	return pCfg, key, resolvedModelName, resp, nil
}

func SetupRoutes(r chi.Router, selector *balancer.Selector, logger *log.Logger, reg *provider.Registry, cfg *config.Config, store store.RuntimeStore) {
	handler := NewChatCompletionsHandler(selector, logger, reg, cfg, store)
	r.With(AuthMiddleware(cfg.APIKeys)).Post("/v1/chat/completions", handler.Handle)

	embeddingsHandler := NewEmbeddingsHandler(selector, logger, reg, cfg, store)
	r.With(AuthMiddleware(cfg.APIKeys)).Post("/v1/embeddings", embeddingsHandler.Handle)

	SetupModelsRoute(r, cfg)
}
