package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/user/coo-llm/internal/api"
	"github.com/user/coo-llm/internal/balancer"
	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/log"
	"github.com/user/coo-llm/internal/provider"
	"github.com/user/coo-llm/internal/store"
)

var version = "1.2.28"

func main() {
	configPath := flag.String("config", "dummy", "path to config file (optional, uses env vars if not set)")
	versionFlag := flag.Bool("version", false, "show version")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	// Load config from CONFIG_PATH env var, flag, or default config.local.yaml or config.yaml
	cfgPath := ""
	if *configPath != "dummy" {
		cfgPath = *configPath
	}
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		cfgPath = envPath
	}
	if cfgPath == "" {
		// Try configs/config.local.yaml first, then configs/config.yaml, then config.local.yaml, config.yaml
		if _, err := os.Stat("configs/config.local.yaml"); err == nil {
			cfgPath = "configs/config.local.yaml"
		} else if _, err := os.Stat("configs/config.yaml"); err == nil {
			cfgPath = "configs/config.yaml"
		} else if _, err := os.Stat("config.local.yaml"); err == nil {
			cfgPath = "config.local.yaml"
		} else if _, err := os.Stat("config.yaml"); err == nil {
			cfgPath = "config.yaml"
		}
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

	// Override admin credentials with env vars
	if adminID := os.Getenv("ADMIN_ID"); adminID != "" {
		cfg.Server.WebUI.AdminID = adminID
	}
	if adminPassword := os.Getenv("ADMIN_PASSWORD"); adminPassword != "" {
		cfg.Server.WebUI.AdminPassword = adminPassword
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
	default:
		// Default to SQL database (PostgreSQL or SQLite based on connection string)
		sqlStore, err := store.NewSQLStore(cfg.Storage.Runtime.Addr, logger.GetLogger())
		if err != nil {
			fmt.Printf("Failed to connect to SQL database: %v\n", err)
			os.Exit(1)
		}
		runtimeStore = sqlStore
	}

	// Create store provider wrapper
	configStore := store.NewSimpleConfigStore(runtimeStore)
	storeProvider := store.NewStoreProviderWrapper(runtimeStore, configStore)

	// Save masked config to store for sharing between instances
	safeCfg := config.MaskSensitiveConfig(cfg)
	if err := configStore.SaveConfig(safeCfg); err != nil {
		fmt.Printf("Warning: Failed to save config to store: %v\n", err)
	}

	// Init selector
	selector := balancer.NewSelector(cfg, storeProvider, logger)

	// Setup router
	r := chi.NewRouter()

	// API routes
	apiRouter := chi.NewRouter()
	// CORS middleware for API routes
	apiRouter.Use(api.CORSMiddleware(cfg))
	api.SetupRoutes(apiRouter, selector, logger, reg, cfg, runtimeStore)
	api.SetupAdminRoutes(apiRouter, cfg, storeProvider, selector, logger)
	r.Mount("/api", apiRouter)

	// Web UI (serve after API routes so it acts as fallback)
	if cfg.Server.WebUI.Enabled {
		var buildDir string

		// If custom WebUIPath is set, use it
		if cfg.Server.WebUI.WebUIPath != "" {
			if _, err := os.Stat(cfg.Server.WebUI.WebUIPath); err == nil {
				buildDir = cfg.Server.WebUI.WebUIPath
			} else {
				fmt.Printf("Warning: Custom Web UI path not found: %s, falling back to built-in\n", cfg.Server.WebUI.WebUIPath)
			}
		}

		// If not set or not found, try built-in paths
		if buildDir == "" {
			possiblePaths := []string{
				"./webui/build",                      // From current dir
				filepath.Join(".", "webui", "build"), // Explicit
				"./build",
			}

			// From executable dir
			if execPath, err := os.Executable(); err == nil {
				execDir := filepath.Dir(execPath)
				possiblePaths = append(possiblePaths,
					filepath.Join(execDir, "..", "webui", "build"), // bin/../webui/build
					filepath.Join(execDir, "webui", "build"),       // bin/webui/build
					filepath.Join(execDir, "build"),                // bin/build (copied)
				)
			}

			for _, path := range possiblePaths {
				if _, err := os.Stat(path); err == nil {
					buildDir = path
					break
				}
			}
		}

		if buildDir == "" {
			fmt.Printf("Warning: Web UI build directory not found, disabling Web UI\n")
		} else {
			fmt.Printf("Serving Web UI from: %s\n", buildDir)
			fileServer := http.FileServer(http.Dir(buildDir))

			// Serve Web UI routes on the main router (not under /api)
			api.SetupWebUIRoutes(r, fileServer)
			fmt.Printf("Web UI enabled at /ui\n")
		}
	}

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
