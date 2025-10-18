package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/user/coo-llm/internal/balancer"
	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/log"
	"github.com/user/coo-llm/internal/provider"
	"github.com/user/coo-llm/internal/store"
)

type EmbeddingsHandler struct {
	selector *balancer.Selector
	logger   *log.Logger
	reg      *provider.Registry
	cfg      *config.Config
	store    store.RuntimeStore
}

type EmbeddingsRequest struct {
	Model          string      `json:"model"`
	Input          interface{} `json:"input"` // string or []string
	User           string      `json:"user,omitempty"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     int         `json:"dimensions,omitempty"`
}

type EmbeddingsResponse struct {
	Object string      `json:"object"`
	Data   []Embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  Usage       `json:"usage"`
}

type Embedding struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type Usage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

func NewEmbeddingsHandler(selector *balancer.Selector, logger *log.Logger, reg *provider.Registry, cfg *config.Config, store store.RuntimeStore) *EmbeddingsHandler {
	return &EmbeddingsHandler{
		selector: selector,
		logger:   logger,
		reg:      reg,
		cfg:      cfg,
		store:    store,
	}
}

func (h *EmbeddingsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Parse request
	var req EmbeddingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Model == "" {
		http.Error(w, "model is required", http.StatusBadRequest)
		return
	}
	if req.Input == nil {
		http.Error(w, "input is required", http.StatusBadRequest)
		return
	}

	// Get client key from context
	// clientKey := r.Context().Value("client_key").(string)

	// Select provider and key
	pCfg, key, modelName, err := h.selector.SelectBest(req.Model)
	if err != nil {
		// TODO: Fix logger - h.logger.GetLogger().Error().Err(err).Str("model", req.Model).Msg("Failed to select provider")
		http.Error(w, "No provider available for model", http.StatusServiceUnavailable)
		return
	}

	// Convert input to strings
	var inputs []string
	switch v := req.Input.(type) {
	case string:
		inputs = []string{v}
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				inputs = append(inputs, str)
			}
		}
	case []string:
		inputs = v
	default:
		http.Error(w, "input must be string or array of strings", http.StatusBadRequest)
		return
	}

	if len(inputs) == 0 {
		http.Error(w, "input cannot be empty", http.StatusBadRequest)
		return
	}

	// Get provider
	prov, err := h.reg.Get(pCfg.ID)
	if err != nil {
		// TODO: Fix logger - Provider not found: %s, pCfg.ID
		http.Error(w, "Provider not available", http.StatusServiceUnavailable)
		return
	}

	// Create provider request
	providerReq := &provider.EmbeddingsRequest{
		Model: modelName,
		Input: inputs,
		User:  req.User,
		Params: map[string]any{
			"encoding_format": req.EncodingFormat,
			"dimensions":      req.Dimensions,
		},
	}

	// Make request
	resp, err := prov.CreateEmbeddings(r.Context(), providerReq)
	if err != nil {
		// TODO: Fix logger - Embeddings request failed: %v for provider %s, err, pCfg.ID

		// Update error metrics
		if key != nil {
			h.selector.UpdateUsage(pCfg.ID, key.ID, "errors", 1)
		}

		http.Error(w, fmt.Sprintf("Provider error: %v", err), http.StatusInternalServerError)
		return
	}

	// Update usage metrics
	latency := time.Since(startTime).Milliseconds()
	if key != nil {
		h.selector.UpdateUsage(pCfg.ID, key.ID, "input_tokens", float64(resp.Usage.PromptTokens))
		h.selector.UpdateUsage(pCfg.ID, key.ID, "tokens", float64(resp.Usage.TotalTokens))
		h.selector.UpdateUsage(pCfg.ID, key.ID, "latency", float64(latency))
		h.selector.UpdateUsage(pCfg.ID, key.ID, "requests", 1)
	}

	// Convert to OpenAI format
	openaiResp := EmbeddingsResponse{
		Object: "list",
		Data:   make([]Embedding, len(resp.Embeddings)),
		Model:  req.Model,
		Usage: Usage{
			PromptTokens: resp.Usage.PromptTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}

	for i, embedding := range resp.Embeddings {
		openaiResp.Data[i] = Embedding{
			Object:    "embedding",
			Embedding: embedding,
			Index:     i,
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openaiResp)

	// TODO: Fix logger - Embeddings request completed for model %s, input_count %d, latency %dms, client %s, req.Model, len(inputs), latency, clientKey
}
