package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/user/truckllm/internal/config"
)

type ClaudeProvider struct {
	cfg *config.Provider
}

func NewClaudeProvider(cfg *config.Provider) *ClaudeProvider {
	return &ClaudeProvider{cfg: cfg}
}

func (p *ClaudeProvider) Name() string {
	return p.cfg.ID
}

func (p *ClaudeProvider) Generate(ctx context.Context, req *Request) (*Response, error) {
	apiKey := req.APIKey
	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided for provider %s", p.Name())
	}

	url := p.cfg.BaseURL + "/v1/messages"

	reqBody, err := json.Marshal(req.Input)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	start := time.Now()
	resp, err := client.Do(httpReq)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &Response{Err: err, Latency: latency}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Response{Err: err, Latency: latency}, nil
	}

	return &Response{
		RawResponse: body,
		HTTPCode:    resp.StatusCode,
		Latency:     latency,
	}, nil
}

func (p *ClaudeProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"claude-3-opus-20240229", "claude-3-sonnet-20240229"}, nil
}
