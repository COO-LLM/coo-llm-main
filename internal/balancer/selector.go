package balancer

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/user/truckllm/internal/config"
	"github.com/user/truckllm/internal/store"
)

type Selector struct {
	cfg   *config.Config
	store store.RuntimeStore
}

func NewSelector(cfg *config.Config, store store.RuntimeStore) *Selector {
	return &Selector{cfg: cfg, store: store}
}

func (s *Selector) SelectBest(model string) (*config.Provider, *config.Key, error) {
	// Resolve provider from model alias
	providerID, modelName := s.resolveModel(model)
	if providerID == "" {
		return nil, nil, fmt.Errorf("model not found: %s", model)
	}

	var pCfg *config.Provider
	for i := range s.cfg.Providers {
		if s.cfg.Providers[i].ID == providerID {
			pCfg = &s.cfg.Providers[i]
			break
		}
	}
	if pCfg == nil {
		return nil, nil, fmt.Errorf("provider not found: %s", providerID)
	}

	// Select best key based on strategy
	key, err := s.selectKey(pCfg, modelName)
	if err != nil {
		return nil, nil, err
	}

	return pCfg, key, nil
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
	switch s.cfg.Policy.Strategy {
	case "round_robin":
		return s.selectRoundRobin(pCfg)
	case "least_error":
		return s.selectLeastError(pCfg)
	case "hybrid":
		return s.selectHybrid(pCfg, model)
	default:
		return s.selectRoundRobin(pCfg)
	}
}

func (s *Selector) selectRoundRobin(pCfg *config.Provider) (*config.Key, error) {
	if len(pCfg.Keys) == 0 {
		return nil, fmt.Errorf("no keys available")
	}
	// Simple round robin, in real impl use store to track
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &pCfg.Keys[r.Intn(len(pCfg.Keys))], nil
}

func (s *Selector) selectLeastError(pCfg *config.Provider) (*config.Key, error) {
	// Implement based on error rates from store
	// For now, return first
	if len(pCfg.Keys) == 0 {
		return nil, fmt.Errorf("no keys available")
	}
	return &pCfg.Keys[0], nil
}

func (s *Selector) selectHybrid(pCfg *config.Provider, model string) (*config.Key, error) {
	// Implement hybrid scoring
	best := &pCfg.Keys[0]
	minScore := math.MaxFloat64

	for _, key := range pCfg.Keys {
		score := s.calculateScore(pCfg.ID, &key, model)
		if score < minScore {
			minScore = score
			best = &key
		}
	}
	return best, nil
}

func (s *Selector) calculateScore(providerID string, key *config.Key, model string) float64 {
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
