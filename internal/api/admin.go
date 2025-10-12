package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/user/truckllm/internal/config"
)

type AdminHandler struct {
	cfg *config.Config
}

func NewAdminHandler(cfg *config.Config) *AdminHandler {
	return &AdminHandler{cfg: cfg}
}

func (h *AdminHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.cfg)
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
