package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CustomProvider struct {
	cfg LLMConfig
}

func NewCustomProvider(cfg LLMConfig) *CustomProvider {
	return &CustomProvider{cfg: cfg}
}

func (p *CustomProvider) Name() string {
	return string(ProviderCustom)
}

func (p *CustomProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
	apiKey := p.cfg.NextAPIKey()
	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided for custom provider")
	}

	baseURL := p.cfg.BaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("base URL required for custom provider")
	}

	// Generic request body - can be customized
	reqBody := map[string]any{
		"prompt":     req.Prompt,
		"max_tokens": req.MaxTokens,
	}
	for k, v := range req.Params {
		reqBody[k] = v
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("custom provider API error: %s", string(body))
	}

	// Generic response parsing - assume JSON with "text" and "tokens_used"
	var customResp struct {
		Text         string `json:"text"`
		TokensUsed   int    `json:"tokens_used"`
		InputTokens  int    `json:"input_tokens,omitempty"`
		OutputTokens int    `json:"output_tokens,omitempty"`
	}
	if err := json.Unmarshal(body, &customResp); err != nil {
		return nil, err
	}

	inputTokens := customResp.InputTokens
	outputTokens := customResp.OutputTokens
	if inputTokens == 0 && outputTokens == 0 && customResp.TokensUsed > 0 {
		// Estimate if not provided
		inputTokens = customResp.TokensUsed / 2
		outputTokens = customResp.TokensUsed / 2
	}

	return &LLMResponse{
		Text:         customResp.Text,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TokensUsed:   customResp.TokensUsed,
		FinishReason: "stop",
	}, nil
}

func (p *CustomProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{p.cfg.Model}, nil
}
