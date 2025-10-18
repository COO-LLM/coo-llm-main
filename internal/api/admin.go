package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/user/coo-llm/internal/balancer"
	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/log"
	"github.com/user/coo-llm/internal/store"
)

type AdminHandler struct {
	cfg      *config.Config
	store    store.StoreProvider
	selector *balancer.Selector
	logger   *log.Logger
}

func NewAdminHandler(cfg *config.Config, store store.StoreProvider, selector *balancer.Selector, logger *log.Logger) *AdminHandler {
	return &AdminHandler{cfg: cfg, store: store, selector: selector, logger: logger}
}

func (h *AdminHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	// Load config from store
	cfg, err := h.store.LoadConfig()
	if err != nil {
		// Fallback to in-memory config if store fails
		cfg = h.maskSensitiveConfig(h.cfg)
	} else {
		// Store already has masked config, but ensure it's fully masked
		cfg = h.maskSensitiveConfig(cfg)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

func (h *AdminHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var newCfg config.Config
	if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate the config
	if err := config.ValidateConfig(&newCfg); err != nil {
		http.Error(w, fmt.Sprintf("Invalid config: %v", err), http.StatusBadRequest)
		return
	}

	// Prevent updating sensitive fields (provider API keys, admin key, etc.)
	// Only allow updating public fields like api_keys, policy, etc.
	currentCfg, err := h.store.LoadConfig()
	if err != nil {
		http.Error(w, "Failed to load current config", http.StatusInternalServerError)
		return
	}

	// Merge: keep sensitive fields from current, update public from new
	updatedCfg := *currentCfg
	updatedCfg.APIKeys = newCfg.APIKeys
	updatedCfg.ModelAliases = newCfg.ModelAliases
	updatedCfg.Policy = newCfg.Policy
	// Add other public fields as needed, but not LLMProviders (to protect keys)

	// Save updated config to store
	if err := h.store.SaveConfig(&updatedCfg); err != nil {
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Config updated successfully"})
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

func (h *AdminHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var policyUpdate struct {
		Algorithm string `json:"algorithm"`
		Priority  string `json:"priority"`
		Cache     *struct {
			Enabled    bool  `json:"enabled"`
			TTLSeconds int64 `json:"ttl_seconds"`
		} `json:"cache,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&policyUpdate); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Normalize algorithm (convert dash to underscore for internal consistency)
	policyUpdate.Algorithm = strings.ReplaceAll(policyUpdate.Algorithm, "-", "_")

	// Validate algorithm
	validAlgorithms := map[string]bool{
		"random":       true,
		"round_robin":  true,
		"least_loaded": true,
		"weighted":     true,
	}
	if !validAlgorithms[policyUpdate.Algorithm] {
		http.Error(w, fmt.Sprintf("Invalid algorithm: %s. Must be one of: random, round_robin, least_loaded, weighted", policyUpdate.Algorithm), http.StatusBadRequest)
		return
	}

	// Normalize priority (convert dash to underscore for internal consistency)
	policyUpdate.Priority = strings.ReplaceAll(policyUpdate.Priority, "-", "_")

	// Validate priority
	validPriorities := map[string]bool{
		"latency":      true,
		"cost":         true,
		"availability": true,
		"quality":      true,
		"balanced":     true,
	}
	if !validPriorities[policyUpdate.Priority] {
		http.Error(w, fmt.Sprintf("Invalid priority: %s. Must be one of: latency, cost, availability, quality, balanced", policyUpdate.Priority), http.StatusBadRequest)
		return
	}

	// Validate cache TTL if cache is provided
	if policyUpdate.Cache != nil && policyUpdate.Cache.Enabled && policyUpdate.Cache.TTLSeconds < 1 {
		http.Error(w, "Cache TTL must be at least 1 second", http.StatusBadRequest)
		return
	}

	// Load current config
	currentCfg, err := h.store.LoadConfig()
	if err != nil {
		logger := h.logger.GetLogger()
		logger.Error().Err(err).Msg("Failed to load current config")
		http.Error(w, "Failed to load current config", http.StatusInternalServerError)
		return
	}

	// Update policy fields
	currentCfg.Policy.Algorithm = policyUpdate.Algorithm
	currentCfg.Policy.Priority = policyUpdate.Priority

	if policyUpdate.Cache != nil {
		currentCfg.Policy.Cache.Enabled = policyUpdate.Cache.Enabled
		currentCfg.Policy.Cache.TTLSeconds = policyUpdate.Cache.TTLSeconds
	}

	// Save updated config to store
	if err := h.store.SaveConfig(currentCfg); err != nil {
		logger := h.logger.GetLogger()
		logger.Error().Err(err).Msg("Failed to save config")
		http.Error(w, "Failed to save config", http.StatusInternalServerError)
		return
	}

	// Update in-memory config (h.cfg is a pointer, so we can update it directly)
	h.cfg.Policy.Algorithm = policyUpdate.Algorithm
	h.cfg.Policy.Priority = policyUpdate.Priority
	if policyUpdate.Cache != nil {
		h.cfg.Policy.Cache.Enabled = policyUpdate.Cache.Enabled
		h.cfg.Policy.Cache.TTLSeconds = policyUpdate.Cache.TTLSeconds
	}

	logger := h.logger.GetLogger()
	logger.Info().
		Str("algorithm", policyUpdate.Algorithm).
		Str("priority", policyUpdate.Priority).
		Bool("cache_enabled", currentCfg.Policy.Cache.Enabled).
		Int64("cache_ttl", currentCfg.Policy.Cache.TTLSeconds).
		Msg("Policy updated successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Policy updated successfully",
		"policy": map[string]interface{}{
			"algorithm": currentCfg.Policy.Algorithm,
			"priority":  currentCfg.Policy.Priority,
			"cache": map[string]interface{}{
				"enabled":     currentCfg.Policy.Cache.Enabled,
				"ttl_seconds": currentCfg.Policy.Cache.TTLSeconds,
			},
		},
	})
}

func (h *AdminHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "latency" // default
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end int64
	if startStr != "" {
		start, _ = strconv.ParseInt(startStr, 10, 64)
	} else {
		start = time.Now().Add(-time.Hour).Unix() // last hour
	}

	if endStr != "" {
		end, _ = strconv.ParseInt(endStr, 10, 64)
	} else {
		end = time.Now().Unix()
	}

	points, err := h.store.GetMetrics(name, map[string]string{}, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":   name,
		"start":  start,
		"end":    end,
		"points": points,
	})
}

func (h *AdminHandler) GetClientStats(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end int64
	if startStr != "" {
		start, _ = strconv.ParseInt(startStr, 10, 64)
	} else {
		start = time.Now().Add(-time.Hour * 24).Unix() // last 24 hours
	}

	if endStr != "" {
		end, _ = strconv.ParseInt(endStr, 10, 64)
	} else {
		end = time.Now().Unix()
	}

	// Get all client keys from config
	clientStats := make(map[string]interface{})

	for _, apiKey := range h.cfg.APIKeys {
		key := apiKey.Key
		clientData := map[string]interface{}{
			"total": map[string]float64{
				"queries": 0,
				"tokens":  0,
				"cost":    0,
			},
			"by_provider": make(map[string]map[string]float64),
		}

		// Get metrics for this client
		latencyPoints, _ := h.store.GetMetrics("latency", map[string]string{"client_key": key}, start, end)
		tokenPoints, _ := h.store.GetMetrics("tokens", map[string]string{"client_key": key}, start, end)
		costPoints, _ := h.store.GetMetrics("cost", map[string]string{"client_key": key}, start, end)

		// Create maps for tokens and cost by timestamp
		tokenMap := make(map[int64]float64)
		for _, p := range tokenPoints {
			tokenMap[p.Timestamp] = p.Value
		}
		costMap := make(map[int64]float64)
		for _, p := range costPoints {
			costMap[p.Timestamp] = p.Value
		}

		// Aggregate by provider
		for _, p := range latencyPoints {
			provider := p.Tags["provider"]
			if provider == "" {
				provider = "unknown"
			}
			if _, ok := clientData["by_provider"].(map[string]map[string]float64)[provider]; !ok {
				clientData["by_provider"].(map[string]map[string]float64)[provider] = map[string]float64{
					"queries": 0,
					"tokens":  0,
					"cost":    0,
				}
			}
			clientData["by_provider"].(map[string]map[string]float64)[provider]["queries"]++
			clientData["by_provider"].(map[string]map[string]float64)[provider]["tokens"] += tokenMap[p.Timestamp]
			clientData["by_provider"].(map[string]map[string]float64)[provider]["cost"] += costMap[p.Timestamp]
		}

		// Sum totals
		for _, providerStats := range clientData["by_provider"].(map[string]map[string]float64) {
			clientData["total"].(map[string]float64)["queries"] += providerStats["queries"]
			clientData["total"].(map[string]float64)["tokens"] += providerStats["tokens"]
			clientData["total"].(map[string]float64)["cost"] += providerStats["cost"]
		}

		clientStats[key] = clientData
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"start":        start,
		"end":          end,
		"client_stats": clientStats,
	})
}

func (h *AdminHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	groupByStr := r.URL.Query().Get("group_by")
	if groupByStr == "" {
		groupByStr = "client_key" // default
	}
	groupBy := strings.Split(groupByStr, ",")

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end int64
	if startStr != "" {
		start, _ = strconv.ParseInt(startStr, 10, 64)
	} else {
		start = time.Now().Add(-time.Hour * 24).Unix() // last 24 hours
	}

	if endStr != "" {
		end, _ = strconv.ParseInt(endStr, 10, 64)
	} else {
		end = time.Now().Unix()
	}

	// Filters
	filters := make(map[string]string)
	for _, g := range groupBy {
		if val := r.URL.Query().Get(g); val != "" {
			filters[g] = val
		}
	}
	// Additional filters
	if client := r.URL.Query().Get("client_key"); client != "" {
		filters["client_key"] = client
	}
	if provider := r.URL.Query().Get("provider"); provider != "" {
		filters["provider"] = provider
	}
	if key := r.URL.Query().Get("key"); key != "" {
		filters["key"] = key
	}

	// Get all metrics types
	metricTypes := []string{"latency", "tokens", "cost"}
	allPoints := make(map[string][]store.MetricPoint)
	for _, mt := range metricTypes {
		points, _ := h.store.GetMetrics(mt, filters, start, end)
		allPoints[mt] = points
	}

	// Aggregate by group_by dimensions
	stats := h.aggregateByGroups(allPoints, groupBy)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"start":    start,
		"end":      end,
		"group_by": groupBy,
		"filters":  filters,
		"stats":    stats,
	})
}

func (h *AdminHandler) aggregateByGroups(allPoints map[string][]store.MetricPoint, groupBy []string) map[string]interface{} {
	// Create a map to hold aggregated data
	aggregated := make(map[string]interface{})

	// Collect all unique combinations of group keys
	groupKeys := make(map[string]bool)
	for _, points := range allPoints {
		for _, p := range points {
			keyParts := make([]string, len(groupBy))
			for i, g := range groupBy {
				keyParts[i] = p.Tags[g]
			}
			groupKey := strings.Join(keyParts, "|")
			groupKeys[groupKey] = true
		}
	}

	// For each group key, aggregate metrics
	for groupKey := range groupKeys {
		keyParts := strings.Split(groupKey, "|")
		current := aggregated
		for i, part := range keyParts {
			if part == "" {
				part = "unknown"
			}
			if i == len(keyParts)-1 {
				// Leaf node: aggregate stats
				stats := map[string]float64{
					"queries": 0,
					"tokens":  0,
					"cost":    0,
				}
				// Count queries from latency points
				for _, p := range allPoints["latency"] {
					if h.matchesGroup(p, groupBy, keyParts) {
						stats["queries"]++
					}
				}
				// Sum tokens and cost
				for _, p := range allPoints["tokens"] {
					if h.matchesGroup(p, groupBy, keyParts) {
						stats["tokens"] += p.Value
					}
				}
				for _, p := range allPoints["cost"] {
					if h.matchesGroup(p, groupBy, keyParts) {
						stats["cost"] += p.Value
					}
				}
				current[part] = stats
			} else {
				if current[part] == nil {
					current[part] = make(map[string]interface{})
				}
				current = current[part].(map[string]interface{})
			}
		}
	}

	return aggregated
}

func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AdminID  string `json:"admin_id"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.AdminID == h.cfg.Server.WebUI.AdminID && req.Password == h.cfg.Server.WebUI.AdminPassword {
		// Generate JWT token
		token, err := h.generateJWTToken(req.AdminID)
		if err != nil {
			http.Error(w, "Failed to generate token", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"token": token,
			"user": map[string]string{
				"id":       req.AdminID,
				"username": req.AdminID,
				"role":     "admin",
			},
		})
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}
}

func (h *AdminHandler) generateJWTToken(adminID string) (string, error) {
	// Use admin API key as JWT secret, or a configured secret
	secret := h.cfg.Server.AdminAPIKey
	if secret == "" {
		secret = "default-jwt-secret-change-in-production"
	}

	claims := jwt.MapClaims{
		"admin_id": adminID,
		"type":     "webui",
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (h *AdminHandler) matchesGroup(p store.MetricPoint, groupBy []string, keyParts []string) bool {
	for i, g := range groupBy {
		if p.Tags[g] != keyParts[i] {
			return false
		}
	}
	return true
}

// Client management methods
func (h *AdminHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClientID         string   `json:"client_id"`
		APIKey           string   `json:"api_key"`
		Description      string   `json:"description"`
		AllowedProviders []string `json:"allowed_providers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.ClientID == "" || req.APIKey == "" {
		http.Error(w, "client_id and api_key are required", http.StatusBadRequest)
		return
	}

	err := h.store.CreateClient(req.ClientID, req.APIKey, req.Description, req.AllowedProviders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "created", "client_id": req.ClientID})
}

func (h *AdminHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	clients, err := h.store.ListClients()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"clients": clients})
}

func (h *AdminHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "client_id")
	if clientID == "" {
		http.Error(w, "client_id parameter required", http.StatusBadRequest)
		return
	}

	client, err := h.store.GetClient(clientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}

func (h *AdminHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "client_id")
	if clientID == "" {
		http.Error(w, "client_id parameter required", http.StatusBadRequest)
		return
	}

	var req struct {
		Description      string   `json:"description"`
		AllowedProviders []string `json:"allowed_providers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.store.UpdateClient(clientID, req.Description, req.AllowedProviders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "updated", "client_id": clientID})
}

func (h *AdminHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "client_id")
	if clientID == "" {
		http.Error(w, "client_id parameter required", http.StatusBadRequest)
		return
	}

	err := h.store.DeleteClient(clientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "client_id": clientID})
}

// Enhanced metrics methods
func (h *AdminHandler) GetClientMetrics(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "client_id")
	if clientID == "" {
		http.Error(w, "client_id parameter required", http.StatusBadRequest)
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end int64
	if startStr != "" {
		start, _ = strconv.ParseInt(startStr, 10, 64)
	} else {
		start = time.Now().Add(-time.Hour * 24).Unix() // last 24 hours
	}

	if endStr != "" {
		end, _ = strconv.ParseInt(endStr, 10, 64)
	} else {
		end = time.Now().Unix()
	}

	metrics, err := h.store.GetClientMetrics(clientID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *AdminHandler) GetProviderMetrics(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "provider_id")
	if providerID == "" {
		http.Error(w, "provider_id parameter required", http.StatusBadRequest)
		return
	}

	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end int64
	if startStr != "" {
		start, _ = strconv.ParseInt(startStr, 10, 64)
	} else {
		start = time.Now().Add(-time.Hour * 24).Unix() // last 24 hours
	}

	if endStr != "" {
		end, _ = strconv.ParseInt(endStr, 10, 64)
	} else {
		end = time.Now().Unix()
	}

	metrics, err := h.store.GetProviderMetrics(providerID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (h *AdminHandler) GetGlobalMetrics(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end int64
	if startStr != "" {
		start, _ = strconv.ParseInt(startStr, 10, 64)
	} else {
		start = time.Now().Add(-time.Hour * 24).Unix() // last 24 hours
	}

	if endStr != "" {
		end, _ = strconv.ParseInt(endStr, 10, 64)
	} else {
		end = time.Now().Unix()
	}

	metrics, err := h.store.GetGlobalMetrics(start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func SetupAdminRoutes(r chi.Router, cfg *config.Config, store store.StoreProvider, selector *balancer.Selector, logger *log.Logger) {
	handler := NewAdminHandler(cfg, store, selector, logger)

	// Login endpoint (no auth required)
	r.Post("/login", handler.Login)

	// Admin routes with auth
	adminRouter := chi.NewRouter()
	adminRouter.Use(AdminAuthMiddleware(cfg.Server.AdminAPIKey, cfg))
	adminRouter.Use(RateLimitMiddleware())
	adminRouter.Use(AuditLogMiddleware(logger))

	adminRouter.Get("/v1/config", handler.GetConfig)
	adminRouter.Post("/v1/config", handler.UpdateConfig)
	adminRouter.Post("/v1/config/validate", handler.ValidateConfig)
	adminRouter.Put("/v1/config/policy", handler.UpdatePolicy)
	adminRouter.Get("/v1/metrics", handler.GetMetrics)
	adminRouter.Get("/v1/clients", handler.GetClientStats)
	adminRouter.Get("/v1/stats", handler.GetStats)

	// Client management
	adminRouter.Post("/v1/clients", handler.CreateClient)
	adminRouter.Get("/v1/clients/list", handler.ListClients)
	adminRouter.Get("/v1/clients/{client_id}", handler.GetClient)
	adminRouter.Put("/v1/clients/{client_id}", handler.UpdateClient)
	adminRouter.Delete("/v1/clients/{client_id}", handler.DeleteClient)

	// Enhanced metrics
	adminRouter.Get("/v1/metrics/clients/{client_id}", handler.GetClientMetrics)
	adminRouter.Get("/v1/metrics/providers/{provider_id}", handler.GetProviderMetrics)
	adminRouter.Get("/v1/metrics/global", handler.GetGlobalMetrics)

	// Mount admin router
	r.Mount("/admin", adminRouter)
}

func SetupWebUIRoutes(r chi.Router, fileServer http.Handler) {
	// Handle /ui root path (redirect to /ui/)
	r.Get("/ui", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/ui/", http.StatusMovedPermanently)
	})

	// Create a file server that strips the /ui prefix
	uiFileServer := http.StripPrefix("/ui", fileServer)

	// Serve all static files under /ui/ - BEFORE catch-all
	r.Get("/ui/*", func(w http.ResponseWriter, req *http.Request) {
		// Check if this is a static file request (not a SPA route)
		path := strings.TrimPrefix(req.URL.Path, "/ui/")
		if strings.Contains(path, ".") {
			// This looks like a static file (has extension)
			uiFileServer.ServeHTTP(w, req)
		} else {
			// This is a SPA route, serve index.html
			req.URL.Path = "/"
			fileServer.ServeHTTP(w, req)
		}
	})
}

// Rate limiting for admin endpoints
var requestCounts = make(map[string]int)
var requestTimestamps = make(map[string]time.Time)
var rateLimitMutex sync.Mutex

func RateLimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simple rate limiting: 100 requests per minute per IP
			clientIP := r.RemoteAddr
			if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
				clientIP = strings.Split(forwarded, ",")[0]
			}

			rateLimitMutex.Lock()
			now := time.Now()
			key := clientIP

			// Reset counter if more than a minute has passed
			if timestamp, exists := requestTimestamps[key]; exists && now.Sub(timestamp) > time.Minute {
				requestCounts[key] = 0
			}

			requestCounts[key]++
			requestTimestamps[key] = now

			count := requestCounts[key]
			rateLimitMutex.Unlock()

			if count > 100 {
				http.Error(w, `{"error": {"message": "Rate limit exceeded", "type": "rate_limit_error"}}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func AuditLogMiddleware(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Simple audit logging using fmt for now
			fmt.Printf("[AUDIT] %s %s from %s\n", r.Method, r.URL.Path, r.RemoteAddr)

			next.ServeHTTP(w, r)

			// Log the response
			duration := time.Since(start)
			fmt.Printf("[AUDIT] %s %s completed in %dms\n", r.Method, r.URL.Path, duration.Milliseconds())
		})
	}
}

func AdminAuthMiddleware(adminKey string, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow OPTIONS requests for CORS preflight
			if r.Method == "OPTIONS" {
				next.ServeHTTP(w, r)
				return
			}

			auth := r.Header.Get("Authorization")
			if auth == "" {
				// Add CORS headers even for error responses
				if cfg.Server.CORS.Enabled {
					if len(cfg.Server.CORS.AllowedOrigins) == 1 && cfg.Server.CORS.AllowedOrigins[0] == "*" {
						w.Header().Set("Access-Control-Allow-Origin", "*")
					}
					if cfg.Server.CORS.AllowCredentials {
						w.Header().Set("Access-Control-Allow-Credentials", "true")
					}
					if len(cfg.Server.CORS.AllowedMethods) > 0 {
						w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.Server.CORS.AllowedMethods, ", "))
					}
					if len(cfg.Server.CORS.AllowedHeaders) > 0 {
						w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.Server.CORS.AllowedHeaders, ", "))
					}
				}
				http.Error(w, `{"error": {"message": "Authorization header required", "type": "authentication_error"}}`, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, `{"error": {"message": "Bearer token required", "type": "authentication_error"}}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(auth, "Bearer ")

			// Check static admin API key
			if token == adminKey {
				next.ServeHTTP(w, r)
				return
			}

			// Check JWT token
			if validateJWTToken(token, adminKey, cfg) {
				next.ServeHTTP(w, r)
				return
			}

			// Check legacy webui token (for backward compatibility)
			if strings.HasPrefix(token, "webui-") {
				if validateWebUIToken(token, cfg) {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"error": {"message": "Invalid admin token", "type": "authentication_error"}}`, http.StatusUnauthorized)
		})
	}
}

func validateWebUIToken(token string, cfg *config.Config) bool {
	parts := strings.Split(token, "-")
	if len(parts) != 3 || parts[0] != "webui" {
		return false
	}

	adminID := parts[1]
	timestampStr := parts[2]

	// Verify admin ID matches
	if adminID != cfg.Server.WebUI.AdminID {
		return false
	}

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return false
	}

	// Check if token is not expired (24 hours)
	tokenTime := time.Unix(timestamp, 0)
	if time.Since(tokenTime) > 24*time.Hour {
		return false
	}

	return true
}

func CORSMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !cfg.Server.CORS.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Set CORS headers
			origin := r.Header.Get("Origin")
			if len(cfg.Server.CORS.AllowedOrigins) == 1 && cfg.Server.CORS.AllowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if containsString(cfg.Server.CORS.AllowedOrigins, origin) || containsString(cfg.Server.CORS.AllowedOrigins, "*") {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			if cfg.Server.CORS.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if len(cfg.Server.CORS.AllowedMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.Server.CORS.AllowedMethods, ", "))
			}

			if len(cfg.Server.CORS.AllowedHeaders) > 0 {
				if len(cfg.Server.CORS.AllowedHeaders) == 1 && cfg.Server.CORS.AllowedHeaders[0] == "*" {
					// Copy allowed headers from request
					if requestHeaders := r.Header.Get("Access-Control-Request-Headers"); requestHeaders != "" {
						w.Header().Set("Access-Control-Allow-Headers", requestHeaders)
					} else {
						w.Header().Set("Access-Control-Allow-Headers", "*")
					}
				} else {
					w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.Server.CORS.AllowedHeaders, ", "))
				}
			}

			if cfg.Server.CORS.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.Server.CORS.MaxAge))
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func validateJWTToken(tokenString, adminKey string, cfg *config.Config) bool {
	// Use admin API key as JWT secret, or a configured secret
	secret := adminKey
	if secret == "" {
		secret = "default-jwt-secret-change-in-production"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Verify admin_id matches
		if adminID, ok := claims["admin_id"].(string); ok {
			if adminID != cfg.Server.WebUI.AdminID {
				return false
			}
		} else {
			return false
		}

		// Verify token type
		if tokenType, ok := claims["type"].(string); ok {
			if tokenType != "webui" {
				return false
			}
		} else {
			return false
		}

		return true
	}

	return false
}
