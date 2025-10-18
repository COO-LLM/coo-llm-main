package provider

import (
	"context"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

type GrokProvider struct {
	cfg    *LLMConfig
	client *openai.Client
}

func NewGrokProvider(cfg *LLMConfig) *GrokProvider {
	config := openai.DefaultConfig(cfg.APIKey())
	config.BaseURL = "https://api.x.ai/v1"
	client := openai.NewClientWithConfig(config)
	return &GrokProvider{cfg: cfg, client: client}
}

func (p *GrokProvider) Name() string {
	return string(ProviderGrok)
}

func (p *GrokProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
	// Convert to OpenAI format
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		role := msg["role"].(string)
		content := msg["content"].(string)
		messages[i] = openai.ChatCompletionMessage{
			Role:    role,
			Content: content,
		}
	}

	request := openai.ChatCompletionRequest{
		Model:    p.cfg.Model,
		Messages: messages,
	}

	if req.MaxTokens > 0 {
		request.MaxTokens = req.MaxTokens
	}

	resp, err := p.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("grok API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from grok")
	}

	return &LLMResponse{
		Text:         resp.Choices[0].Message.Content,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
		TokensUsed:   resp.Usage.TotalTokens,
		FinishReason: string(resp.Choices[0].FinishReason),
	}, nil
}

func (p *GrokProvider) GenerateStream(ctx context.Context, req *LLMRequest) (<-chan *LLMStreamResponse, error) {
	streamChan := make(chan *LLMStreamResponse, 10)

	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		role := msg["role"].(string)
		content := msg["content"].(string)
		messages[i] = openai.ChatCompletionMessage{
			Role:    role,
			Content: content,
		}
	}

	request := openai.ChatCompletionRequest{
		Model:    p.cfg.Model,
		Messages: messages,
		Stream:   true,
	}

	if req.MaxTokens > 0 {
		request.MaxTokens = req.MaxTokens
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("grok stream API error: %w", err)
	}

	go func() {
		defer close(streamChan)
		defer stream.Close()

		for {
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					streamChan <- &LLMStreamResponse{Done: true}
					return
				}
				// Send error as a response
				streamChan <- &LLMStreamResponse{Text: fmt.Sprintf("Error: %v", err), Done: true}
				return
			}

			if len(response.Choices) > 0 {
				choice := response.Choices[0]
				streamChan <- &LLMStreamResponse{
					Text:         choice.Delta.Content,
					FinishReason: string(choice.FinishReason),
					Done:         false,
				}
			}
		}
	}()

	return streamChan, nil
}

func (p *GrokProvider) CreateEmbeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	// Grok uses OpenAI-compatible API, try embeddings endpoint
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

		config := openai.DefaultConfig(currentKey)
		config.BaseURL = "https://api.x.ai/v1"
		p.client = openai.NewClientWithConfig(config)

		modelName := p.cfg.Model
		if req.Model != "" {
			modelName = req.Model
		}

		embedReq := openai.EmbeddingRequest{
			Input: req.Input,
			Model: openai.EmbeddingModel(modelName),
			User:  req.User,
		}

		resp, err := p.client.CreateEmbeddings(ctx, embedReq)
		if err == nil && len(resp.Data) > 0 {
			// Update usage - estimate tokens based on input length
			totalTokens := 0
			for _, input := range req.Input {
				totalTokens += len(input) / 4 // Rough estimate
			}
			p.cfg.UpdateUsage(len(resp.Data), totalTokens)

			embeddings := make([]Embedding, len(resp.Data))
			for i, data := range resp.Data {
				// Convert []float32 to []float64
				embedding := make([]float64, len(data.Embedding))
				for j, val := range data.Embedding {
					embedding[j] = float64(val)
				}
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

		// If error and not last attempt, continue to next key
		if attempt == maxRetries-1 {
			// Last attempt failed
			if err != nil {
				return nil, fmt.Errorf("Grok embeddings API error after %d attempts: %w", maxRetries, err)
			}
			return nil, fmt.Errorf("no embeddings response from Grok after %d attempts", maxRetries)
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

func (p *GrokProvider) ListModels(ctx context.Context) ([]string, error) {
	// Grok typically supports grok-beta and similar models
	return []string{"grok-beta"}, nil
}
