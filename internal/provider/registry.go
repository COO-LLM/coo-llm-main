package provider

import (
	"fmt"
	"sync"

	"github.com/user/truckllm/internal/config"
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
	for _, pCfg := range cfg.Providers {
		var p Provider
		switch pCfg.ID {
		case "openai":
			p = NewOpenAIProvider(&pCfg)
		case "gemini":
			p = NewGeminiProvider(&pCfg)
		case "claude":
			p = NewClaudeProvider(&pCfg)
		default:
			return fmt.Errorf("unsupported provider: %s", pCfg.ID)
		}
		r.Register(p)
	}
	return nil
}
