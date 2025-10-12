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
	models := make([]map[string]interface{}, 0)
	for alias, model := range h.cfg.ModelAliases {
		models = append(models, map[string]interface{}{
			"id":       alias,
			"object":   "model",
			"created":  1677610602,
			"owned_by": model,
		})
	}

	resp := map[string]interface{}{
		"object": "list",
		"data":   models,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func SetupModelsRoute(r chi.Router, cfg *config.Config) {
	handler := NewModelsHandler(cfg)
	r.Get("/v1/models", handler.Handle)
}
