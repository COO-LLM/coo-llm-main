package balancer

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/log"
	"github.com/user/coo-llm/internal/store"
)

type Selector struct {
	cfg    *config.Config
	store  store.StoreProvider
	logger *log.Logger
}

func NewSelector(cfg *config.Config, store store.StoreProvider, logger *log.Logger) *Selector {
	return &Selector{cfg: cfg, store: store, logger: logger}
}

// getCurrentPolicy loads the current policy from store, with fallback to config
func (s *Selector) getCurrentPolicy() config.Policy {
	// Try to load from store first
	storedCfg, err := s.store.LoadConfig()
	if err == nil && storedCfg != nil {
		return storedCfg.Policy
	}

	// Fallback to static config
	if s.logger != nil {
		logger := s.logger.GetLogger()
		logger.Warn().Err(err).Msg("Failed to load policy from store, using static config")
	}
	return s.cfg.Policy
}

func (s *Selector) SelectBest(model string) (*config.Provider, *config.Key, string, error) {
	// Resolve provider from model alias
	providerID, modelName := s.resolveModel(model)
	if providerID == "" {
		return nil, nil, "", fmt.Errorf("model not found: %s", model)
	}

	// Try LLMProviders first (new format)
	for i := range s.cfg.LLMProviders {
		if s.cfg.LLMProviders[i].ID == providerID {
			// Convert LLMProvider to Provider format for backward compatibility
			lp := s.cfg.LLMProviders[i]
			sessionLimit := lp.Limits.SessionLimit
			if sessionLimit == 0 {
				sessionLimit = lp.Limits.TokensPerMin * 60
			}
			sessionType := lp.Limits.SessionType
			if sessionType == "" {
				sessionType = "1h"
			}
			keys := make([]config.Key, len(lp.APIKeys))
			for j, apiKey := range lp.APIKeys {
				// Use hash of API key as stable ID for consistency
				h := sha256.Sum256([]byte(apiKey))
				keyID := fmt.Sprintf("%s-%x", lp.ID, h[:8])
				keys[j] = config.Key{
					ID:                keyID,
					Secret:            apiKey,
					LimitReqPerMin:    lp.Limits.ReqPerMin,
					LimitTokensPerMin: lp.Limits.TokensPerMin,
					SessionLimit:      sessionLimit,
					SessionType:       sessionType,
				}
			}
			llmProvider := &config.Provider{
				ID:      lp.ID,
				Name:    lp.Name,
				BaseURL: lp.BaseURL,
				Limits:  lp.Limits,
				Pricing: lp.Pricing,
				Keys:    keys,
			}
			key, err := s.selectKey(llmProvider, modelName)
			if err != nil {
				return nil, nil, "", err
			}
			return llmProvider, key, modelName, nil
		}
	}

	// Fallback to legacy Providers
	for i := range s.cfg.Providers {
		if s.cfg.Providers[i].ID == providerID {
			pCfg := &s.cfg.Providers[i]
			key, err := s.selectKey(pCfg, modelName)
			if err != nil {
				return nil, nil, "", err
			}
			return pCfg, key, modelName, nil
		}
	}

	return nil, nil, "", fmt.Errorf("provider not found: %s", providerID)
}

func (s *Selector) resolveModel(model string) (string, string) {
	// Check if model is in provider:model format
	if colonIndex := strings.Index(model, ":"); colonIndex != -1 {
		return model[:colonIndex], model[colonIndex+1:]
	}

	// Check model aliases
	if alias, ok := s.cfg.ModelAliases[model]; ok {
		// Parse alias like "openai:gpt-4o"
		parts := strings.Split(alias, ":")
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
	}

	// Default to openai if no alias
	return "openai", model
}

func (s *Selector) selectKey(pCfg *config.Provider, model string) (*config.Key, error) {
	policy := s.getCurrentPolicy()
	switch policy.Algorithm {
	case "round_robin":
		return s.selectRoundRobin(pCfg)
	case "least_loaded":
		return s.selectLeastLoaded(pCfg)
	case "hybrid":
		return s.selectHybrid(pCfg, model, policy)
	default:
		return s.selectRoundRobin(pCfg)
	}
}

func (s *Selector) isRateLimited(pCfg *config.Provider, key *config.Key) bool {
	// Check requests per minute with burst allowance (10% over)
	reqUsage, _ := s.store.GetUsageInWindow(pCfg.ID, key.ID, "req", 60) // 60 seconds window
	if reqUsage >= float64(key.LimitReqPerMin)*1.1 {
		return true
	}

	// Check tokens per minute with burst allowance (10% over)
	tokenUsage, _ := s.store.GetUsageInWindow(pCfg.ID, key.ID, "tokens", 60) // 60 seconds window
	if tokenUsage >= float64(key.LimitTokensPerMin)*1.1 {
		return true
	}

	// Check session limit with burst allowance (10% over)
	if key.SessionLimit > 0 && key.SessionType != "" {
		sessionSeconds := s.parseSessionType(key.SessionType)
		if sessionSeconds > 0 {
			tokenUsageSession, _ := s.store.GetUsageInWindow(pCfg.ID, key.ID, "tokens", sessionSeconds)
			if tokenUsageSession >= float64(key.SessionLimit)*1.1 {
				return true
			}
		}
	}

	return false
}

func (s *Selector) selectRoundRobin(pCfg *config.Provider) (*config.Key, error) {
	if len(pCfg.Keys) == 0 {
		return nil, fmt.Errorf("no keys available")
	}

	// Try to find a non-rate-limited key
	availableKeys := make([]*config.Key, 0, len(pCfg.Keys))
	for i := range pCfg.Keys {
		if !s.isRateLimited(pCfg, &pCfg.Keys[i]) {
			availableKeys = append(availableKeys, &pCfg.Keys[i])
		}
	}

	// If no non-rate-limited keys, use any key (allow bursting)
	if len(availableKeys) == 0 {
		availableKeys = make([]*config.Key, len(pCfg.Keys))
		for i := range pCfg.Keys {
			availableKeys[i] = &pCfg.Keys[i]
		}
	}

	// Simple round robin
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return availableKeys[r.Intn(len(availableKeys))], nil
}

func (s *Selector) selectLeastLoaded(pCfg *config.Provider) (*config.Key, error) {
	// Select key with lowest total usage, preferring non-rate-limited keys
	if len(pCfg.Keys) == 0 {
		return nil, fmt.Errorf("no keys available")
	}

	var best *config.Key
	minUsage := math.MaxFloat64

	// First pass: try to find non-rate-limited keys
	for i := range pCfg.Keys {
		key := &pCfg.Keys[i]
		if s.isRateLimited(pCfg, key) {
			continue
		}
		usage, _ := s.store.GetUsage(pCfg.ID, key.ID, "tokens")
		if usage < minUsage {
			minUsage = usage
			best = key
		}
	}

	// If no non-rate-limited keys found, use any key
	if best == nil {
		for i := range pCfg.Keys {
			key := &pCfg.Keys[i]
			usage, _ := s.store.GetUsage(pCfg.ID, key.ID, "tokens")
			if usage < minUsage {
				minUsage = usage
				best = key
			}
		}
	}

	return best, nil
}

func (s *Selector) selectHybrid(pCfg *config.Provider, model string, policy config.Policy) (*config.Key, error) {
	// Implement hybrid scoring, preferring non-rate-limited keys
	var best *config.Key
	minScore := math.MaxFloat64

	// First pass: try to find non-rate-limited keys
	for i := range pCfg.Keys {
		key := &pCfg.Keys[i]
		if s.isRateLimited(pCfg, key) {
			continue
		}
		score := s.calculateScore(pCfg, key, model, policy)
		if score < minScore {
			minScore = score
			best = key
		}
	}

	// If no non-rate-limited keys found, use any key
	if best == nil {
		for i := range pCfg.Keys {
			key := &pCfg.Keys[i]
			score := s.calculateScore(pCfg, key, model, policy)
			if score < minScore {
				minScore = score
				best = key
			}
		}
	}

	return best, nil
}

func (s *Selector) calculateScore(pCfg *config.Provider, key *config.Key, _ string, policy config.Policy) float64 {
	w := policy.HybridWeights

	providerID := pCfg.ID
	reqUsage, _ := s.store.GetUsage(providerID, key.ID, "req")
	tokenUsage, _ := s.store.GetUsage(providerID, key.ID, "tokens")
	errorScore, _ := s.store.GetUsage(providerID, key.ID, "errors")
	latency, _ := s.store.GetUsage(providerID, key.ID, "latency")

	// Estimate cost based on average tokens (assume 1000 tokens per request for simplicity)
	avgTokens := 1000.0
	estimatedCost := (pCfg.Pricing.InputTokenCost + pCfg.Pricing.OutputTokenCost) * avgTokens / 1000

	score := w.ReqRatio*reqUsage + w.TokenRatio*tokenUsage + w.ErrorScore*errorScore + w.Latency*latency + w.CostRatio*estimatedCost

	// Prioritize providers with higher MaxTokens (subtract to lower score)
	if pCfg.Limits.MaxTokens > 0 {
		score -= float64(pCfg.Limits.MaxTokens) / 1000.0 * 0.1 // Small weight for MaxTokens
	}

	return score
}

func (s *Selector) UpdateUsage(providerID, keyID, metric string, delta float64) {
	s.store.IncrementUsage(providerID, keyID, metric, delta)
}

func (s *Selector) GetCache(key string) (string, error) {
	return s.store.GetCache(key)
}

func (s *Selector) SetCache(key, value string, ttlSeconds int64) error {
	return s.store.SetCache(key, value, ttlSeconds)
}

func (s *Selector) SelectKeyForProvider(pCfg *config.Provider, model string) (*config.Key, error) {
	return s.selectKey(pCfg, model)
}

func (s *Selector) GetRecommendedKey(pCfg *config.Provider, model string) *config.Key {
	policy := s.getCurrentPolicy()
	var best *config.Key
	minScore := math.MaxFloat64
	for i := range pCfg.Keys {
		key := &pCfg.Keys[i]
		score := s.calculateScore(pCfg, key, model, policy)
		if score < minScore {
			minScore = score
			best = key
		}
	}
	return best
}

func (s *Selector) parseSessionType(sessionType string) int64 {
	// Parse session type like "5m", "5h", "1d" to seconds
	if len(sessionType) < 2 {
		return 3600 // default 1h
	}
	numStr := sessionType[:len(sessionType)-1]
	unit := sessionType[len(sessionType)-1]
	num := 1
	if n, err := strconv.Atoi(numStr); err == nil {
		num = n
	}
	switch unit {
	case 'm':
		return int64(num * 60)
	case 'h':
		return int64(num * 3600)
	case 'd':
		return int64(num * 86400)
	default:
		return 3600
	}
}
