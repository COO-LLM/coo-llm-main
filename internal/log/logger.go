package log

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/user/truckllm/internal/config"
)

type Logger struct {
	cfg    *config.Logging
	logger zerolog.Logger
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	ReqID     string `json:"req_id"`
	LatencyMS int64  `json:"latency_ms"`
	Status    int    `json:"status"`
	Tokens    int    `json:"tokens"`
	Error     string `json:"error,omitempty"`
}

func NewLogger(cfg *config.Logging) *Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	if cfg.File.Enabled {
		file, err := os.OpenFile(cfg.File.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			logger = zerolog.New(file).With().Timestamp().Logger()
		}
	}
	return &Logger{cfg: cfg, logger: logger}
}

func (l *Logger) LogRequest(ctx context.Context, entry *LogEntry) {
	entry.Timestamp = time.Now().Format(time.RFC3339)
	data, _ := json.Marshal(entry)
	l.logger.Info().RawJSON("entry", data).Msg("request")

	// Send to providers if configured
	for _, p := range l.cfg.Providers {
		go l.sendToProvider(p, entry)
	}
}

func (l *Logger) sendToProvider(p config.LogProvider, entry *LogEntry) {
	switch p.Type {
	case "http":
		l.sendHTTP(p, entry)
	default:
		l.logger.Info().Str("provider", p.Name).Interface("entry", entry).Msg("sending to log provider")
	}
}

func (l *Logger) sendHTTP(p config.LogProvider, entry *LogEntry) {
	// Implement HTTP POST to p.Endpoint
	l.logger.Info().Str("endpoint", p.Endpoint).Interface("entry", entry).Msg("sending log via HTTP")
}
