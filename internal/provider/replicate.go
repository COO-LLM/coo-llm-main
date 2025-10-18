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

type ReplicateProvider struct {
	cfg    *LLMConfig
	client *http.Client
}

type ReplicatePredictionRequest struct {
	Version string                 `json:"version"`
	Input   map[string]interface{} `json:"input"`
}

type ReplicatePredictionResponse struct {
	ID      string                 `json:"id"`
	Status  string                 `json:"status"`
	Output  interface{}            `json:"output"`
	Error   string                 `json:"error"`
	Logs    string                 `json:"logs"`
	Metrics map[string]interface{} `json:"metrics"`
}

func NewReplicateProvider(cfg *LLMConfig) *ReplicateProvider {
	return &ReplicateProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: 60 * time.Second}, // Replicate can be slow
	}
}

func (p *ReplicateProvider) Name() string {
	return string(ProviderReplicate)
}

func (p *ReplicateProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
	// Replicate uses model versions, not model names
	// We need to map model names to Replicate model versions
	modelVersion := p.mapModelToVersion(req.Model)
	if modelVersion == "" {
		modelVersion = p.cfg.Model // Use configured model as version
	}

	// Prepare input for Replicate
	input := map[string]interface{}{
		"prompt": req.Prompt,
	}

	// Add optional parameters
	if req.MaxTokens > 0 {
		input["max_tokens"] = req.MaxTokens
	}
	if temp, ok := req.Params["temperature"].(float64); ok {
		input["temperature"] = temp
	}
	if topP, ok := req.Params["top_p"].(float64); ok {
		input["top_p"] = topP
	}

	replicateReq := ReplicatePredictionRequest{
		Version: modelVersion,
		Input:   input,
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
		reqBody, err := json.Marshal(replicateReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST",
			"https://api.replicate.com/v1/predictions", bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Authorization", "Token "+currentKey)
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := p.client.Do(httpReq)
		if err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("Replicate API error after %d attempts: %w", maxRetries, err)
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("Replicate API error: %s - %s", resp.Status, string(body))
			}
			continue
		}

		var predictionResp ReplicatePredictionResponse
		if err := json.NewDecoder(resp.Body).Decode(&predictionResp); err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("failed to decode response: %w", err)
			}
			continue
		}

		// For synchronous responses, check if output is ready
		if predictionResp.Status == "succeeded" && predictionResp.Output != nil {
			// Extract text from output
			var text string
			switch output := predictionResp.Output.(type) {
			case string:
				text = output
			case []interface{}:
				// Some models return array of strings
				for _, item := range output {
					if str, ok := item.(string); ok {
						text += str
					}
				}
			}

			if text != "" {
				// Estimate tokens
				inputTokens := len(req.Prompt) / 4
				outputTokens := len(text) / 4
				totalTokens := inputTokens + outputTokens

				// Update usage
				p.cfg.UpdateUsage(1, totalTokens)

				return &LLMResponse{
					Text:         text,
					InputTokens:  inputTokens,
					OutputTokens: outputTokens,
					TokensUsed:   totalTokens,
					FinishReason: "stop",
				}, nil
			}
		}

		// If prediction is not ready, we would need to poll for results
		// For simplicity, return error for async predictions
		if attempt == maxRetries-1 {
			return nil, fmt.Errorf("prediction not ready or failed: %s", predictionResp.Error)
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

func (p *ReplicateProvider) GenerateStream(ctx context.Context, req *LLMRequest) (<-chan *LLMStreamResponse, error) {
	// Replicate doesn't have built-in streaming support like OpenAI
	// Return non-streaming response as single chunk
	streamChan := make(chan *LLMStreamResponse, 1)

	go func() {
		defer close(streamChan)
		resp, err := p.Generate(ctx, req)
		if err != nil {
			streamChan <- &LLMStreamResponse{Text: fmt.Sprintf("Error: %v", err), Done: true}
			return
		}
		streamChan <- &LLMStreamResponse{Text: resp.Text, FinishReason: resp.FinishReason, Done: true}
	}()

	return streamChan, nil
}

func (p *ReplicateProvider) CreateEmbeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	// Replicate may not have dedicated embedding models
	// Most Replicate models are for text generation, not embeddings
	return nil, fmt.Errorf("embeddings not supported by Replicate provider - use Cohere or OpenAI for embeddings")
}

func (p *ReplicateProvider) ListModels(ctx context.Context) ([]string, error) {
	// Replicate hosts many models, return some popular ones
	// Note: These are model version identifiers, not names
	return []string{
		"meta/llama-2-70b-chat",
		"mistralai/mistral-7b-instruct-v0.1",
		"replicate/llama-2-70b-chat",
		"stability-ai/stable-diffusion",
		"cjwbw/anything-v3-better-vae",
		"andreasjansson/stable-diffusion-animation",
		"riffusion/riffusion",
		"openai/whisper",
		"daanelson/minigpt-4",
		"lucataco/animate-diff",
	}, nil
}

// mapModelToVersion maps common model names to Replicate version IDs
func (p *ReplicateProvider) mapModelToVersion(model string) string {
	// This is a simplified mapping - in practice, you'd need to maintain
	// a comprehensive mapping of model names to Replicate version IDs
	switch model {
	case "llama-2-70b-chat":
		return "02e509c789964a7ea8736978a43525956ef40397be9033abf9fcd80d8580583"
	case "mistral-7b-instruct":
		return "83b6a56e7c828e667f21fd596c338fd4f0039b46bcfa18d973e8e70e3519"
	case "stable-diffusion":
		return "db21e45d3f7023abc2a46ee38a23973f6dce16bb082a930b0c49861f96"
	default:
		// Return the model name as-is if no mapping found
		return model
	}
}
