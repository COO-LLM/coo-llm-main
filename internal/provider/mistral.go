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

type MistralProvider struct {
	cfg    *LLMConfig
	client *http.Client
}

type MistralChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

type MistralChatResponse struct {
	ID      string    `json:"id"`
	Object  string    `json:"object"`
	Created int64     `json:"created"`
	Model   string    `json:"model"`
	Choices []Choice  `json:"choices"`
	Usage   UsageInfo `json:"usage"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func NewMistralProvider(cfg *LLMConfig) *MistralProvider {
	return &MistralProvider{
		cfg:    cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *MistralProvider) Name() string {
	return string(ProviderMistral)
}

func (p *MistralProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
	// Convert messages to Mistral format
	var messages []Message
	if len(req.Messages) > 0 {
		messages = make([]Message, len(req.Messages))
		for i, msg := range req.Messages {
			role, _ := msg["role"].(string)
			content, _ := msg["content"].(string)
			messages[i] = Message{
				Role:    role,
				Content: content,
			}
		}
	} else {
		// Fallback to single message
		messages = []Message{
			{Role: "user", Content: req.Prompt},
		}
	}

	modelName := p.cfg.Model
	if req.Model != "" {
		modelName = req.Model
	}

	mistralReq := MistralChatRequest{
		Model:     modelName,
		Messages:  messages,
		MaxTokens: req.MaxTokens,
	}

	// Add custom params
	if temp, ok := req.Params["temperature"].(float64); ok {
		mistralReq.Temperature = temp
	}
	if topP, ok := req.Params["top_p"].(float64); ok {
		mistralReq.TopP = topP
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
		reqBody, err := json.Marshal(mistralReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST",
			"https://api.mistral.ai/v1/chat/completions", bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+currentKey)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("Mistral API error after %d attempts: %w", maxRetries, err)
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("Mistral API error: %s - %s", resp.Status, string(body))
			}
			continue
		}

		var mistralResp MistralChatResponse
		if err := json.NewDecoder(resp.Body).Decode(&mistralResp); err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("failed to decode response: %w", err)
			}
			continue
		}

		if len(mistralResp.Choices) > 0 {
			// Update usage
			p.cfg.UpdateUsage(1, mistralResp.Usage.TotalTokens)

			return &LLMResponse{
				Text:         mistralResp.Choices[0].Message.Content,
				InputTokens:  mistralResp.Usage.PromptTokens,
				OutputTokens: mistralResp.Usage.CompletionTokens,
				TokensUsed:   mistralResp.Usage.TotalTokens,
				FinishReason: mistralResp.Choices[0].FinishReason,
			}, nil
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

func (p *MistralProvider) GenerateStream(ctx context.Context, req *LLMRequest) (<-chan *LLMStreamResponse, error) {
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

type MistralEmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type MistralEmbedResponse struct {
	Object string                 `json:"object"`
	Data   []MistralEmbeddingData `json:"data"`
	Usage  MistralEmbeddingUsage  `json:"usage"`
}

type MistralEmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type MistralEmbeddingUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

func (p *MistralProvider) CreateEmbeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	// Mistral supports embeddings
	// Create HTTP request for embeddings

	type MistralEmbedResponse struct {
		Object string                 `json:"object"`
		Data   []MistralEmbeddingData `json:"data"`
		Usage  MistralEmbeddingUsage  `json:"usage"`
	}

	type MistralEmbeddingData struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	}

	type MistralEmbeddingUsage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	}

	modelName := p.cfg.Model
	if req.Model != "" {
		modelName = req.Model
	}

	embedReq := MistralEmbedRequest{
		Model: modelName,
		Input: req.Input,
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
		reqBody, err := json.Marshal(embedReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST",
			"https://api.mistral.ai/v1/embeddings", bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create embedding request: %w", err)
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+currentKey)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("Mistral embeddings API error after %d attempts: %w", maxRetries, err)
			}
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if attempt == maxRetries-1 {
				return nil, fmt.Errorf("Mistral embeddings API error: %s - %s", resp.Status, string(body))
			}
			continue
		}

		var embedResp MistralEmbedResponse
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
					PromptTokens: embedResp.Usage.PromptTokens,
					TotalTokens:  embedResp.Usage.TotalTokens,
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

func (p *MistralProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{
		"mistral-large-latest",
		"mistral-medium",
		"mistral-small",
		"mistral-7b-instruct",
		"mistral-8x7b-instruct",
		"mistral-embed",
	}, nil
}
