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

type VoyageProvider struct {
	cfg    *LLMConfig
	client *http.Client
}

type VoyageEmbedRequest struct {
	Input     []string `json:"input"`
	Model     string   `json:"model"`
	InputType string   `json:"input_type,omitempty"`
}

type VoyageEmbedResponse struct {
	Object string                `json:"object"`
	Data   []VoyageEmbeddingData `json:"data"`
	Usage  VoyageUsage           `json:"usage"`
}

type VoyageEmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type VoyageUsage struct {
	TotalTokens int `json:"total_tokens"`
}

func NewVoyageProvider(cfg *LLMConfig) *VoyageProvider {
	return &VoyageProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *VoyageProvider) Name() string {
	return string(ProviderVoyage)
}

func (p *VoyageProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
	// Voyage AI primarily focuses on embeddings, not text generation
	// Return error indicating this provider is for embeddings only
	return nil, fmt.Errorf("text generation not supported by Voyage AI provider - use for embeddings only")
}

func (p *VoyageProvider) GenerateStream(ctx context.Context, req *LLMRequest) (<-chan *LLMStreamResponse, error) {
	// Voyage doesn't support streaming text generation
	return nil, fmt.Errorf("streaming not supported by Voyage AI provider")
}

func (p *VoyageProvider) CreateEmbeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	// Voyage AI is specialized in embeddings
	modelName := p.cfg.Model
	if req.Model != "" {
		modelName = req.Model
	}

	voyageReq := VoyageEmbedRequest{
		Input:     req.Input,
		Model:     modelName,
		InputType: "document", // Default for document embeddings
	}

	// Retry with different keys if fail (max 3 attempts)
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Select least loaded key for first attempt, then rotate on retry
		var currentKey string
		if attempt == 0 {
			currentKey = p.cfg.SelectLeastLoadedKey()
		} else {
			currentKey = p.cfg.NextAPIKey()
		}
		if currentKey == "" {
			return nil, fmt.Errorf("no API key available")
		}

		// Create HTTP request
		reqBody, err := json.Marshal(voyageReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST",
			"https://api.voyageai.com/v1/embeddings", bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+currentKey)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("Voyage AI embeddings API error after %d attempts: %w", maxRetries, err)
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("Voyage AI embeddings API error: %s - %s", resp.Status, string(body))
			}
			continue
		}

		var embedResp VoyageEmbedResponse
		if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("failed to decode embedding response: %w", err)
			}
			continue
		}

		if len(embedResp.Data) > 0 {
			// Update usage
			p.cfg.UpdateUsage(len(embedResp.Data), embedResp.Usage.TotalTokens)

			embeddings := make([]Embedding, len(embedResp.Data))
			for i, data := range embedResp.Data {
				embeddings[i] = Embedding(data.Embedding)
			}

			return &EmbeddingsResponse{
				Embeddings: embeddings,
				Usage: TokenUsage{
					PromptTokens: embedResp.Usage.TotalTokens,
					TotalTokens:  embedResp.Usage.TotalTokens,
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

func (p *VoyageProvider) ListModels(ctx context.Context) ([]string, error) {
	// Voyage AI embedding models
	return []string{
		"voyage-3-large",
		"voyage-3.5",
		"voyage-3.5-lite",
		"voyage-code-3",
		"voyage-finance-2",
		"voyage-law-2",
		"voyage-multimodal-3",
	}, nil
}
