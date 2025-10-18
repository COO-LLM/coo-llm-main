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

type CohereProvider struct {
	cfg    *LLMConfig
	client *http.Client
}

type CohereChatRequest struct {
	Model       string          `json:"model"`
	Messages    []CohereMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"p,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

type CohereMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CohereChatResponse struct {
	ID       string         `json:"id"`
	Response CohereResponse `json:"response"`
	Meta     CohereMeta     `json:"meta"`
}

type CohereResponse struct {
	Text         string `json:"text"`
	FinishReason string `json:"finish_reason"`
}

type CohereMeta struct {
	APIVersion CohereAPIVersion `json:"api_version"`
}

type CohereAPIVersion struct {
	Version string `json:"version"`
}

type CohereEmbedRequest struct {
	Model     string   `json:"model"`
	Texts     []string `json:"texts"`
	InputType string   `json:"input_type,omitempty"`
}

type CohereEmbedResponse struct {
	ID         string          `json:"id"`
	Texts      []string        `json:"texts"`
	Embeddings [][]float64     `json:"embeddings"`
	Meta       CohereEmbedMeta `json:"meta"`
}

type CohereEmbedMeta struct {
	APIVersion CohereEmbedAPIVersion `json:"api_version"`
}

type CohereEmbedAPIVersion struct {
	Version string `json:"version"`
}

func NewCohereProvider(cfg *LLMConfig) *CohereProvider {
	return &CohereProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *CohereProvider) Name() string {
	return string(ProviderCohere)
}

func (p *CohereProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
	// Convert messages to Cohere format
	var messages []CohereMessage
	if len(req.Messages) > 0 {
		messages = make([]CohereMessage, len(req.Messages))
		for i, msg := range req.Messages {
			role, _ := msg["role"].(string)
			content, _ := msg["content"].(string)
			messages[i] = CohereMessage{
				Role:    role,
				Content: content,
			}
		}
	} else {
		// Fallback to single message
		messages = []CohereMessage{
			{Role: "user", Content: req.Prompt},
		}
	}

	modelName := p.cfg.Model
	if req.Model != "" {
		modelName = req.Model
	}

	cohereReq := CohereChatRequest{
		Model:     modelName,
		Messages:  messages,
		MaxTokens: req.MaxTokens,
	}

	// Add custom params
	if temp, ok := req.Params["temperature"].(float64); ok {
		cohereReq.Temperature = temp
	}
	if topP, ok := req.Params["top_p"].(float64); ok {
		cohereReq.TopP = topP
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
		reqBody, err := json.Marshal(cohereReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST",
			"https://api.cohere.ai/v1/chat", bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+currentKey)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("Cohere API error after %d attempts: %w", maxRetries, err)
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("Cohere API error: %s - %s", resp.Status, string(body))
			}
			continue
		}

		var cohereResp CohereChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&cohereResp); err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("failed to decode response: %w", err)
			}
			continue
		}

		// Estimate tokens (Cohere doesn't provide exact token counts in response)
		inputTokens := 0
		for _, msg := range messages {
			inputTokens += len(msg.Content) / 4
		}
		outputTokens := len(cohereResp.Response.Text) / 4
		totalTokens := inputTokens + outputTokens

		// Update usage
		p.cfg.UpdateUsage(1, totalTokens)

		return &LLMResponse{
			Text:         cohereResp.Response.Text,
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			TokensUsed:   totalTokens,
			FinishReason: cohereResp.Response.FinishReason,
		}, nil
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

func (p *CohereProvider) GenerateStream(ctx context.Context, req *LLMRequest) (<-chan *LLMStreamResponse, error) {
	// For now, return non-streaming response as single chunk
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

func (p *CohereProvider) CreateEmbeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	// Cohere has excellent embeddings support
	modelName := p.cfg.Model
	if req.Model != "" {
		modelName = req.Model
	}

	cohereReq := CohereEmbedRequest{
		Model:     modelName,
		Texts:     req.Input,
		InputType: "search_document", // Default for document embeddings
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
		reqBody, err := json.Marshal(cohereReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST",
			"https://api.cohere.ai/v1/embed", bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+currentKey)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("Cohere embeddings API error after %d attempts: %w", maxRetries, err)
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("Cohere embeddings API error: %s - %s", resp.Status, string(body))
			}
			continue
		}

		var embedResp CohereEmbedResponse
		if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("failed to decode embedding response: %w", err)
			}
			continue
		}

		if len(embedResp.Embeddings) > 0 {
			// Estimate tokens based on input length
			totalTokens := 0
			for _, input := range req.Input {
				totalTokens += len(input) / 4
			}

			// Update usage
			p.cfg.UpdateUsage(len(embedResp.Embeddings), totalTokens)

			embeddings := make([]Embedding, len(embedResp.Embeddings))
			for i, embedding := range embedResp.Embeddings {
				embeddings[i] = Embedding(embedding)
			}

			return &EmbeddingsResponse{
				Embeddings: embeddings,
				Usage: TokenUsage{
					PromptTokens: totalTokens,
					TotalTokens:  totalTokens,
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

func (p *CohereProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{
		"command-r-plus",
		"command-r",
		"command",
		"command-light",
		"embed-english-v3.0",
		"embed-multilingual-v3.0",
		"embed-english-light-v3.0",
		"embed-multilingual-light-v3.0",
	}, nil
}
