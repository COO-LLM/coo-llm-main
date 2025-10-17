package provider

import (
	"fmt"
	"sync"

	"github.com/user/coo-llm/internal/config"
)

type Registry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

func (r *Registry) RegisterWithID(id string, p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[id] = p
}

func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return p, nil
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

func (r *Registry) LoadFromConfig(cfg *config.Config) error {
	// Load new LLMProviders
	for _, lp := range cfg.LLMProviders {
		llmCfg := LLMConfig{
			Type:    ProviderType(lp.Type),
			APIKeys: lp.APIKeys,
			BaseURL: lp.BaseURL,
			Model:   lp.Model,
			Pricing: lp.Pricing,
			Limits:  lp.Limits,
		}
		p, err := NewLLMProvider(&llmCfg)
		if err != nil {
			return err
		}
		r.RegisterWithID(lp.ID, p)
	}

	// Fallback to legacy Providers if no LLMProviders
	if len(cfg.LLMProviders) == 0 {
		for _, pCfg := range cfg.Providers {
			keys := make([]string, len(pCfg.Keys))
			for i, k := range pCfg.Keys {
				keys[i] = k.Secret
			}
			llmCfg := LLMConfig{
				Type:    ProviderType(pCfg.ID),
				APIKeys: keys,
				BaseURL: pCfg.BaseURL,
				Model:   "gpt-4",
				Pricing: pCfg.Keys[0].Pricing,
				Limits: config.Limits{
					ReqPerMin:    pCfg.Keys[0].LimitReqPerMin,
					TokensPerMin: pCfg.Keys[0].LimitTokensPerMin,
				},
			}
			p, err := NewLLMProvider(&llmCfg)
			if err != nil {
				return err
			}
			r.RegisterWithID(pCfg.ID, p)
		}
	}
	return nil
}
