package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/user/truckllm/internal/api"
	"github.com/user/truckllm/internal/balancer"
	"github.com/user/truckllm/internal/config"
	"github.com/user/truckllm/internal/log"
	"github.com/user/truckllm/internal/provider"
	"github.com/user/truckllm/internal/store"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := config.ValidateConfig(cfg); err != nil {
		fmt.Printf("Invalid config: %v\n", err)
		os.Exit(1)
	}

	// Init registry
	reg := provider.NewRegistry()
	if err := reg.LoadFromConfig(cfg); err != nil {
		fmt.Printf("Failed to load providers: %v\n", err)
		os.Exit(1)
	}

	// Init store
	var runtimeStore store.RuntimeStore
	switch cfg.Storage.Runtime.Type {
	case "redis":
		runtimeStore = store.NewRedisStore(cfg.Storage.Runtime.Addr, cfg.Storage.Runtime.Password)
	case "http":
		runtimeStore = store.NewHTTPStore(cfg.Storage.Runtime.Addr, cfg.Storage.Runtime.APIKey)
	default:
		runtimeStore = store.NewRedisStore("localhost:6379", "")
	}

	// Init selector
	selector := balancer.NewSelector(cfg, runtimeStore)

	// Init logger
	logger := log.NewLogger(&cfg.Logging)

	// Setup router
	r := chi.NewRouter()

	// API routes
	api.SetupRoutes(r, selector, logger, reg, cfg)

	// Metrics
	if cfg.Logging.Prometheus.Enabled {
		r.Handle(cfg.Logging.Prometheus.Endpoint, promhttp.Handler())
	}

	// Admin routes
	fmt.Printf("Starting server on %s\n", cfg.Server.Listen)
	if err := http.ListenAndServe(cfg.Server.Listen, r); err != nil {
		fmt.Printf("Server failed: %v\n", err)
		os.Exit(1)
	}
}
