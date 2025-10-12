package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/user/truckllm/internal/config"
)

type ModelsHandler struct {
	cfg *config.Config
}

func NewModelsHandler(cfg *config.Config) *ModelsHandler {
	return &ModelsHandler{cfg: cfg}
}

func (h *ModelsHandler) Handle(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"object": "list",
		"data":   []any{},
	}

	for alias := range h.cfg.ModelAliases {
		response["data"] = append(response["data"].([]any), map[string]any{
			"id":       alias,
			"object":   "model",
			"created":  1677649963,
			"owned_by": "truckllm",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func SetupModelsRoute(r chi.Router, cfg *config.Config) {
	handler := NewModelsHandler(cfg)
	r.Get("/v1/models", handler.Handle)
}
