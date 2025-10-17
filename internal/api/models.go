package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/user/coo-llm/internal/config"
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
			"owned_by": "coo-llm",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AuthMiddleware checks for Authorization header (Bearer token)
func AuthMiddleware(apiKeyConfigs []config.APIKeyConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, `{"error": {"message": "Missing API key", "type": "authentication_error"}}`, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, `{"error": {"message": "Invalid API key format", "type": "authentication_error"}}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(auth, "Bearer ")

			// If no API keys are configured, accept any token (for development)
			if len(apiKeyConfigs) == 0 {
				ctx := context.WithValue(r.Context(), "allowed_providers", []string{"*"})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Validate against configured API keys and get permissions
			var allowedProviders []string
			valid := false
			for _, keyConfig := range apiKeyConfigs {
				if token == keyConfig.Key {
					valid = true
					allowedProviders = keyConfig.AllowedProviders
					break
				}
			}

			if !valid {
				http.Error(w, `{"error": {"message": "Invalid API key", "type": "authentication_error"}}`, http.StatusUnauthorized)
				return
			}

			// Add allowed providers and API key to context
			ctx := context.WithValue(r.Context(), "allowed_providers", allowedProviders)
			ctx = context.WithValue(ctx, "api_key", token) // Use the token as api_key
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func SetupModelsRoute(r chi.Router, cfg *config.Config) {
	handler := NewModelsHandler(cfg)
	r.With(AuthMiddleware(cfg.APIKeys)).Get("/v1/models", handler.Handle)
}
