package provider

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	cfg    *LLMConfig
	client *genai.Client
}

func NewGeminiProvider(cfg *LLMConfig) *GeminiProvider {
	// Client will be created in Generate method to allow key rotation
	return &GeminiProvider{cfg: cfg, client: nil}
}

func (p *GeminiProvider) Name() string {
	return string(ProviderGemini)
}

func (p *GeminiProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
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

		client, err := genai.NewClient(ctx, option.WithAPIKey(currentKey))
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini client: %w", err)
		}
		p.client = client
		defer p.client.Close()

		modelName := p.cfg.Model
		if req.Model != "" {
			modelName = req.Model
		}
		model := p.client.GenerativeModel(modelName)

		// Set generation config
		maxTokens := int32(req.MaxTokens)
		model.GenerationConfig = genai.GenerationConfig{
			MaxOutputTokens: &maxTokens,
		}
		if temp, ok := req.Params["temperature"].(float64); ok {
			temp32 := float32(temp)
			model.GenerationConfig.Temperature = &temp32
		}
		if topP, ok := req.Params["top_p"].(float64); ok {
			topP32 := float32(topP)
			model.GenerationConfig.TopP = &topP32
		}

		// Handle conversation history
		var resp *genai.GenerateContentResponse
		if len(req.Messages) > 1 {
			// Use chat session for multi-turn conversation
			chat := model.StartChat()

			// Add history (all messages except the last one)
			for i := 0; i < len(req.Messages)-1; i++ {
				msg := req.Messages[i]
				role, _ := msg["role"].(string)
				content, _ := msg["content"].(string)

				if role == "user" {
					chat.SendMessage(ctx, genai.Text(content))
				} else if role == "assistant" {
					// For assistant messages, we need to simulate the response
					// This is a limitation - Gemini doesn't support adding assistant messages to history easily
					// For now, we'll just continue with the last user message
				}
			}

			// Send the last message
			lastMsg := req.Messages[len(req.Messages)-1]
			if content, ok := lastMsg["content"].(string); ok {
				resp, err = chat.SendMessage(ctx, genai.Text(content))
			} else {
				err = fmt.Errorf("no content in last message")
			}
		} else {
			// Single message
			resp, err = model.GenerateContent(ctx, genai.Text(req.Prompt))
		}

		if err == nil && resp != nil && len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
			if ok {
				// Estimate tokens (SDK may not provide exact count)
				inputTokens := len(req.Prompt) / 4
				outputTokens := len(string(text)) / 4
				tokensUsed := inputTokens + outputTokens

				// Update usage
				p.cfg.UpdateUsage(1, tokensUsed)

				return &LLMResponse{
					Text:         string(text),
					InputTokens:  inputTokens,
					OutputTokens: outputTokens,
					TokensUsed:   tokensUsed,
					FinishReason: string(resp.Candidates[0].FinishReason),
				}, nil
			}
		}

		// If error and not last attempt, try next key
		if attempt < maxRetries-1 {
			p.cfg.NextAPIKey()
		} else {
			// Last attempt failed
			if err != nil {
				return nil, fmt.Errorf("Gemini API error after %d attempts: %w", maxRetries, err)
			}
			return nil, fmt.Errorf("no response from Gemini after %d attempts", maxRetries)
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

func (p *GeminiProvider) GenerateStream(ctx context.Context, req *LLMRequest) (<-chan *LLMStreamResponse, error) {
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

func (p *GeminiProvider) CreateEmbeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
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

		client, err := genai.NewClient(ctx, option.WithAPIKey(currentKey))
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini client: %w", err)
		}
		defer client.Close()

		modelName := p.cfg.Model
		if req.Model != "" {
			modelName = req.Model
		}

		// Get embedding model
		embeddingModel := client.EmbeddingModel(modelName)

		// Process each input text
		embeddings := make([]Embedding, len(req.Input))
		totalTokens := 0

		for i, input := range req.Input {
			// Call Gemini embedding API
			resp, err := embeddingModel.EmbedContent(ctx, genai.Text(input))
			if err != nil {
				// If error and not last attempt, try next key
				if attempt < maxRetries-1 {
					p.cfg.NextAPIKey()
					break
				} else {
					return nil, fmt.Errorf("Gemini embeddings API error after %d attempts: %w", maxRetries, err)
				}
			}

			if resp.Embedding == nil {
				return nil, fmt.Errorf("no embedding returned from Gemini")
			}

			// Convert []float32 to []float64
			embedding := make([]float64, len(resp.Embedding.Values))
			for j, val := range resp.Embedding.Values {
				embedding[j] = float64(val)
			}
			embeddings[i] = Embedding(embedding)

			// Estimate tokens (rough approximation)
			totalTokens += len(input) / 4
		}

		// If we get here, all embeddings were successful
		if len(embeddings) == len(req.Input) {
			// Update usage
			p.cfg.UpdateUsage(len(req.Input), totalTokens)

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

func (p *GeminiProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"gemini-1.5-pro", "gemini-1.5-flash", "gemini-1.0-pro"}, nil
}
