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

type GeminiProvider struct {
	cfg *config.Provider
}

func NewGeminiProvider(cfg *config.Provider) *GeminiProvider {
	return &GeminiProvider{cfg: cfg}
}

func (p *GeminiProvider) Name() string {
	return p.cfg.ID
}

func (p *GeminiProvider) Generate(ctx context.Context, req *Request) (*Response, error) {
	apiKey := req.APIKey
	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided for provider %s", p.Name())
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.cfg.BaseURL, req.Model, apiKey)

	reqBody, err := json.Marshal(req.Input)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
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

func (p *GeminiProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"gemini-1.5-pro", "gemini-1.5-flash"}, nil
}
