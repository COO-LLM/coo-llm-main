package balancer

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/store"
)

type Selector struct {
	cfg   *config.Config
	store store.RuntimeStore
}

func NewSelector(cfg *config.Config, store store.RuntimeStore) *Selector {
	return &Selector{cfg: cfg, store: store}
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
			var secret string
			if len(s.cfg.LLMProviders[i].APIKeys) > 0 {
				secret = s.cfg.LLMProviders[i].APIKeys[0]
			}
			llmProvider := &config.Provider{
				ID:      s.cfg.LLMProviders[i].ID,
				BaseURL: s.cfg.LLMProviders[i].BaseURL,
				Keys:    []config.Key{{ID: "default", Secret: secret}},
			}
			return llmProvider, &llmProvider.Keys[0], modelName, nil
		}
	}

	// Fallback to legacy Providers
	for i := range s.cfg.Providers {
		if s.cfg.Providers[i].ID == providerID {
			pCfg := &s.cfg.Providers[i]
			return pCfg, &pCfg.Keys[0], modelName, nil
		}
	}

	return nil, nil, "", fmt.Errorf("provider not found: %s", providerID)

	var pCfg *config.Provider

	// Try LLMProviders first
	for i := range s.cfg.LLMProviders {
		if s.cfg.LLMProviders[i].ID == providerID {
			var secret string
			if len(s.cfg.LLMProviders[i].APIKeys) > 0 {
				secret = s.cfg.LLMProviders[i].APIKeys[0]
			}
			pCfg = &config.Provider{
				ID:      s.cfg.LLMProviders[i].ID,
				BaseURL: s.cfg.LLMProviders[i].BaseURL,
				Keys:    []config.Key{{ID: "default", Secret: secret}},
			}
			break
		}
	}

	// Fallback to legacy Providers
	if pCfg == nil {
		for i := range s.cfg.Providers {
			if s.cfg.Providers[i].ID == providerID {
				pCfg = &s.cfg.Providers[i]
				break
			}
		}
	}

	if pCfg == nil {
		return nil, nil, "", fmt.Errorf("provider not found: %s", providerID)
	}

	// Select best key based on strategy
	key, err := s.selectKey(pCfg, modelName)
	if err != nil {
		return nil, nil, "", err
	}

	return pCfg, key, modelName, nil
}

func (s *Selector) resolveModel(model string) (string, string) {
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
	switch s.cfg.Policy.Algorithm {
	case "round_robin":
		return s.selectRoundRobin(pCfg)
	case "least_loaded":
		return s.selectLeastLoaded(pCfg)
	case "hybrid":
		return s.selectHybrid(pCfg, model)
	default:
		return s.selectRoundRobin(pCfg)
	}
}

func (s *Selector) isRateLimited(pCfg *config.Provider, key *config.Key) bool {
	// Check requests per minute
	reqUsage, _ := s.store.GetUsageInWindow(pCfg.ID, key.ID, "req", 60) // 60 seconds window
	if reqUsage >= float64(key.LimitReqPerMin) {
		return true
	}

	// Check tokens per minute
	tokenUsage, _ := s.store.GetUsageInWindow(pCfg.ID, key.ID, "tokens", 60) // 60 seconds window
	if tokenUsage >= float64(key.LimitTokensPerMin) {
		return true
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

func (s *Selector) selectHybrid(pCfg *config.Provider, model string) (*config.Key, error) {
	// Implement hybrid scoring, preferring non-rate-limited keys
	var best *config.Key
	minScore := math.MaxFloat64

	// First pass: try to find non-rate-limited keys
	for i := range pCfg.Keys {
		key := &pCfg.Keys[i]
		if s.isRateLimited(pCfg, key) {
			continue
		}
		score := s.calculateScore(pCfg.ID, key, model)
		if score < minScore {
			minScore = score
			best = key
		}
	}

	// If no non-rate-limited keys found, use any key
	if best == nil {
		for i := range pCfg.Keys {
			key := &pCfg.Keys[i]
			score := s.calculateScore(pCfg.ID, key, model)
			if score < minScore {
				minScore = score
				best = key
			}
		}
	}

	return best, nil
}

func (s *Selector) calculateScore(providerID string, key *config.Key, _ string) float64 {
	w := s.cfg.Policy.HybridWeights

	reqUsage, _ := s.store.GetUsage(providerID, key.ID, "req")
	tokenUsage, _ := s.store.GetUsage(providerID, key.ID, "tokens")
	errorScore, _ := s.store.GetUsage(providerID, key.ID, "errors")
	latency, _ := s.store.GetUsage(providerID, key.ID, "latency")

	// Estimate cost based on average tokens (assume 1000 tokens per request for simplicity)
	avgTokens := 1000.0
	estimatedCost := (key.Pricing.InputTokenCost + key.Pricing.OutputTokenCost) * avgTokens / 1000

	score := w.ReqRatio*reqUsage + w.TokenRatio*tokenUsage + w.ErrorScore*errorScore + w.Latency*latency + w.CostRatio*estimatedCost
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
