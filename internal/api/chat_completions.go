package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/user/truckllm/internal/balancer"
	"github.com/user/truckllm/internal/config"
	"github.com/user/truckllm/internal/log"
	"github.com/user/truckllm/internal/provider"
)

type ChatCompletionsHandler struct {
	selector *balancer.Selector
	logger   *log.Logger
	reg      *provider.Registry
}

func NewChatCompletionsHandler(selector *balancer.Selector, logger *log.Logger, reg *provider.Registry) *ChatCompletionsHandler {
	return &ChatCompletionsHandler{selector: selector, logger: logger, reg: reg}
}

func (h *ChatCompletionsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	model, ok := req["model"].(string)
	if !ok {
		http.Error(w, "model is required", http.StatusBadRequest)
		return
	}

	pCfg, key, err := h.selector.SelectBest(model)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	prov, err := h.reg.Get(pCfg.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	providerReq := &provider.Request{
		Model:  model,
		Input:  req,
		APIKey: key.Secret,
	}

	resp, err := prov.Generate(r.Context(), providerReq)
	if err != nil {
		// Update error usage
		h.selector.UpdateUsage(pCfg.ID, key.ID, "errors", 1)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update usage
	h.selector.UpdateUsage(pCfg.ID, key.ID, "req", 1)
	h.selector.UpdateUsage(pCfg.ID, key.ID, "tokens", float64(resp.TokensUsed))
	h.selector.UpdateUsage(pCfg.ID, key.ID, "latency", float64(resp.Latency))

	// Log the request
	reqID := fmt.Sprintf("%d", time.Now().UnixNano())
	h.logger.LogRequest(r.Context(), &log.LogEntry{
		Provider:  pCfg.ID,
		Model:     model,
		ReqID:     reqID,
		LatencyMS: resp.Latency,
		Status:    resp.HTTPCode,
		Tokens:    resp.TokensUsed,
		Error:     "",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.HTTPCode)
	w.Write(resp.RawResponse)
}

func SetupRoutes(r chi.Router, selector *balancer.Selector, logger *log.Logger, reg *provider.Registry, cfg *config.Config) {
	handler := NewChatCompletionsHandler(selector, logger, reg)
	r.Post("/v1/chat/completions", handler.Handle)

	modelsHandler := NewModelsHandler(cfg)
	r.Get("/v1/models", modelsHandler.Handle)
}
