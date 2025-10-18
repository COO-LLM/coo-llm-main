package provider

import (
	"context"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

type OpenRouterProvider struct {
	cfg    *LLMConfig
	client *openai.Client
}

func NewOpenRouterProvider(cfg *LLMConfig) *OpenRouterProvider {
	config := openai.DefaultConfig(cfg.APIKey())
	config.BaseURL = "https://openrouter.ai/api/v1"
	client := openai.NewClientWithConfig(config)
	return &OpenRouterProvider{cfg: cfg, client: client}
}

func (p *OpenRouterProvider) Name() string {
	return string(ProviderOpenRouter)
}

func (p *OpenRouterProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
	// Convert messages to OpenAI format
	var messages []openai.ChatCompletionMessage
	if len(req.Messages) > 0 {
		messages = make([]openai.ChatCompletionMessage, len(req.Messages))
		for i, msg := range req.Messages {
			role, _ := msg["role"].(string)
			content, _ := msg["content"].(string)

			var openaiRole string
			switch role {
			case "user":
				openaiRole = openai.ChatMessageRoleUser
			case "assistant":
				openaiRole = openai.ChatMessageRoleAssistant
			case "system":
				openaiRole = openai.ChatMessageRoleSystem
			default:
				openaiRole = openai.ChatMessageRoleUser
			}

			messages[i] = openai.ChatCompletionMessage{
				Role:    openaiRole,
				Content: content,
			}
		}
	} else {
		// Fallback to single message
		messages = []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: req.Prompt},
		}
	}

	modelName := p.cfg.Model
	if req.Model != "" {
		modelName = req.Model
	}
	chatReq := openai.ChatCompletionRequest{
		Model:     modelName,
		Messages:  messages,
		MaxTokens: req.MaxTokens,
	}

	// Add custom params
	if temp, ok := req.Params["temperature"].(float64); ok {
		chatReq.Temperature = float32(temp)
	}
	if topP, ok := req.Params["top_p"].(float64); ok {
		chatReq.TopP = float32(topP)
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

		config := openai.DefaultConfig(currentKey)
		config.BaseURL = "https://openrouter.ai/api/v1"
		p.client = openai.NewClientWithConfig(config)

		resp, err := p.client.CreateChatCompletion(ctx, chatReq)
		if err == nil && len(resp.Choices) > 0 {
			// Update usage
			p.cfg.UpdateUsage(1, resp.Usage.TotalTokens)
			return &LLMResponse{
				Text:         resp.Choices[0].Message.Content,
				InputTokens:  resp.Usage.PromptTokens,
				OutputTokens: resp.Usage.CompletionTokens,
				TokensUsed:   resp.Usage.TotalTokens,
				FinishReason: string(resp.Choices[0].FinishReason),
			}, nil
		}

		// If error and not last attempt, continue to next key
		if attempt == maxRetries-1 {
			// Last attempt failed
			if err != nil {
				return nil, fmt.Errorf("OpenRouter API error after %d attempts: %w", maxRetries, err)
			}
			return nil, fmt.Errorf("no response from OpenRouter after %d attempts", maxRetries)
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

func (p *OpenRouterProvider) GenerateStream(ctx context.Context, req *LLMRequest) (<-chan *LLMStreamResponse, error) {
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
		return nil, fmt.Errorf("openrouter stream API error: %w", err)
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

func (p *OpenRouterProvider) CreateEmbeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	// OpenRouter supports embeddings through OpenAI-compatible API
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
		config.BaseURL = "https://openrouter.ai/api/v1"
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
				return nil, fmt.Errorf("OpenRouter embeddings API error after %d attempts: %w", maxRetries, err)
			}
			return nil, fmt.Errorf("no embeddings response from OpenRouter after %d attempts", maxRetries)
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

func (p *OpenRouterProvider) ListModels(ctx context.Context) ([]string, error) {
	// OpenRouter supports hundreds of models from various providers
	// Return some popular ones
	return []string{
		"anthropic/claude-3.5-sonnet",
		"openai/gpt-4o",
		"openai/gpt-4o-mini",
		"meta-llama/llama-3.2-90b-instruct",
		"google/gemini-pro-1.5",
		"mistralai/mistral-7b-instruct",
		"anthropic/claude-3-haiku",
		"openai/gpt-3.5-turbo",
		"meta-llama/llama-3.1-70b-instruct",
		"anthropic/claude-3-opus",
	}, nil
}
