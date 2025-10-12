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

type OpenAIProvider struct {
	cfg *config.Provider
}

func NewOpenAIProvider(cfg *config.Provider) *OpenAIProvider {
	return &OpenAIProvider{cfg: cfg}
}

func (p *OpenAIProvider) Name() string {
	return p.cfg.ID
}

func (p *OpenAIProvider) Generate(ctx context.Context, req *Request) (*Response, error) {
	apiKey := req.APIKey
	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided for provider %s", p.Name())
	}

	url := p.cfg.BaseURL + "/v1/chat/completions" // Assuming chat completions for now

	reqBody, err := json.Marshal(req.Input)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
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

	tokensUsed := 0
	if resp.StatusCode == 200 {
		var jsonResp map[string]interface{}
		if err := json.Unmarshal(body, &jsonResp); err == nil {
			if usage, ok := jsonResp["usage"].(map[string]interface{}); ok {
				if total, ok := usage["total_tokens"].(float64); ok {
					tokensUsed = int(total)
				}
			}
		}
	}

	return &Response{
		RawResponse: body,
		HTTPCode:    resp.StatusCode,
		Latency:     latency,
		TokensUsed:  tokensUsed,
	}, nil
}

func (p *OpenAIProvider) ListModels(ctx context.Context) ([]string, error) {
	// Implement list models if needed
	return []string{"gpt-4o", "gpt-4", "gpt-3.5-turbo"}, nil
}
