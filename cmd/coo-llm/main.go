package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/user/coo-llm/internal/api"
	"github.com/user/coo-llm/internal/balancer"
	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/log"
	"github.com/user/coo-llm/internal/provider"
	"github.com/user/coo-llm/internal/store"
)

var version = "dev"

func main() {
	configPath := flag.String("config", "dummy", "path to config file (optional, uses env vars if not set)")
	versionFlag := flag.Bool("version", false, "show version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	// Use CONFIG_PATH env var if set, else flag value
	cfgPath := ""
	if *configPath != "dummy" {
		cfgPath = *configPath
	}
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		cfgPath = envPath
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Override port with PORT env var if set (for Railway, etc.)
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Listen = ":" + port
	}

	if err := config.ValidateConfig(cfg); err != nil {
		fmt.Printf("Invalid config: %v\n", err)
		os.Exit(1)
	}

	// Init logger
	logger := log.NewLogger(&cfg.Logging)

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
		runtimeStore = store.NewRedisStore(cfg.Storage.Runtime.Addr, cfg.Storage.Runtime.Password, logger.GetLogger())
	case "http":
		runtimeStore = store.NewHTTPStore(cfg.Storage.Runtime.Addr, cfg.Storage.Runtime.APIKey, logger.GetLogger())
	case "sql":
		sqlStore, err := store.NewSQLStore(cfg.Storage.Runtime.Addr, logger.GetLogger())
		if err != nil {
			fmt.Printf("Failed to connect to SQL database: %v\n", err)
			os.Exit(1)
		}
		runtimeStore = sqlStore
	case "mongodb":
		mongoStore, err := store.NewMongoDBStore(cfg.Storage.Runtime.Addr, cfg.Storage.Runtime.Database, logger.GetLogger())
		if err != nil {
			fmt.Printf("Failed to connect to MongoDB: %v\n", err)
			os.Exit(1)
		}
		runtimeStore = mongoStore
	case "dynamodb":
		dynamoStore, err := store.NewDynamoDBStore(cfg.Storage.Runtime.Addr, cfg.Storage.Runtime.TableUsage, cfg.Storage.Runtime.TableCache, cfg.Storage.Runtime.TableHistory, logger.GetLogger())
		if err != nil {
			fmt.Printf("Failed to connect to DynamoDB: %v\n", err)
			os.Exit(1)
		}
		runtimeStore = dynamoStore
	case "influxdb":
		runtimeStore = store.NewInfluxDBStore(cfg.Storage.Runtime.Addr, cfg.Storage.Runtime.Password, cfg.Storage.Runtime.APIKey, cfg.Storage.Runtime.Database, logger.GetLogger())
	case "memory":
		runtimeStore = store.NewMemoryStore(logger.GetLogger())
	default:
		runtimeStore = store.NewMemoryStore(logger.GetLogger()) // Default to memory
	}

	// Init selector
	selector := balancer.NewSelector(cfg, runtimeStore)

	// Setup router
	r := chi.NewRouter()

	// API routes
	api.SetupAdminRoutes(r, cfg, runtimeStore, selector)
	api.SetupRoutes(r, selector, logger, reg, cfg, runtimeStore)

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
