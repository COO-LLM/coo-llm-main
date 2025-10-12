package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/user/coo-llm/internal/config"
)

type AdminHandler struct {
	cfg *config.Config
}

func NewAdminHandler(cfg *config.Config) *AdminHandler {
	return &AdminHandler{cfg: cfg}
}

func (h *AdminHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	// Return config with sensitive data masked
	safeCfg := h.maskSensitiveConfig(h.cfg)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safeCfg)
}

func (h *AdminHandler) maskSensitiveConfig(cfg *config.Config) *config.Config {
	// Create a copy with sensitive fields masked
	safeCfg := *cfg

	// Mask API keys in providers
	for i := range safeCfg.Providers {
		for j := range safeCfg.Providers[i].Keys {
			if len(safeCfg.Providers[i].Keys[j].Secret) > 4 {
				safeCfg.Providers[i].Keys[j].Secret = safeCfg.Providers[i].Keys[j].Secret[:4] + "****"
			}
		}
	}

	// Mask API keys in llm_providers
	for i := range safeCfg.LLMProviders {
		for j := range safeCfg.LLMProviders[i].APIKeys {
			if len(safeCfg.LLMProviders[i].APIKeys[j]) > 4 {
				safeCfg.LLMProviders[i].APIKeys[j] = safeCfg.LLMProviders[i].APIKeys[j][:4] + "****"
			}
		}
	}

	// Mask admin API key
	if len(safeCfg.Server.AdminAPIKey) > 4 {
		safeCfg.Server.AdminAPIKey = safeCfg.Server.AdminAPIKey[:4] + "****"
	}

	// Mask storage passwords/API keys
	if safeCfg.Storage.Runtime.Password != "" {
		safeCfg.Storage.Runtime.Password = "****"
	}
	if safeCfg.Storage.Runtime.APIKey != "" {
		safeCfg.Storage.Runtime.APIKey = "****"
	}

	return &safeCfg
}

func (h *AdminHandler) ValidateConfig(w http.ResponseWriter, r *http.Request) {
	var cfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := config.ValidateConfig(&cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Config is valid"))
}

func SetupAdminRoutes(r chi.Router, cfg *config.Config) {
	r.Use(middleware.BasicAuth("admin", map[string]string{
		"admin": cfg.Server.AdminAPIKey,
	}))

	handler := NewAdminHandler(cfg)
	r.Get("/admin/v1/config", handler.GetConfig)
	r.Post("/admin/v1/config/validate", handler.ValidateConfig)
}
