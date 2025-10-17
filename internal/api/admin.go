package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/user/coo-llm/internal/balancer"
	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/store"
)

type AdminHandler struct {
	cfg      *config.Config
	store    store.RuntimeStore
	selector *balancer.Selector
}

func NewAdminHandler(cfg *config.Config, store store.RuntimeStore, selector *balancer.Selector) *AdminHandler {
	return &AdminHandler{cfg: cfg, store: store, selector: selector}
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

func (h *AdminHandler) matchesGroup(p store.MetricPoint, groupBy []string, keyParts []string) bool {
	for i, g := range groupBy {
		if p.Tags[g] != keyParts[i] {
			return false
		}
	}
	return true
}

func SetupAdminRoutes(r chi.Router, cfg *config.Config, store store.RuntimeStore, selector *balancer.Selector) {
	adminRouter := chi.NewRouter()
	adminRouter.Use(AdminAuthMiddleware(cfg.Server.AdminAPIKey))

	handler := NewAdminHandler(cfg, store, selector)
	adminRouter.Get("/admin/v1/config", handler.GetConfig)
	adminRouter.Post("/admin/v1/config/validate", handler.ValidateConfig)
	adminRouter.Get("/admin/v1/metrics", handler.GetMetrics)
	adminRouter.Get("/admin/v1/clients", handler.GetClientStats)
	adminRouter.Get("/admin/v1/stats", handler.GetStats)

	r.Mount("/", adminRouter)
}

func AdminAuthMiddleware(adminKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, `{"error": {"message": "Authorization header required", "type": "authentication_error"}}`, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, `{"error": {"message": "Bearer token required", "type": "authentication_error"}}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(auth, "Bearer ")
			if token != adminKey {
				http.Error(w, `{"error": {"message": "Invalid admin token", "type": "authentication_error"}}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
