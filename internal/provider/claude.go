package provider

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeProvider struct {
	cfg    *LLMConfig
	client anthropic.Client
}

func NewClaudeProvider(cfg *LLMConfig) *ClaudeProvider {
	// Client will be created in Generate method to allow key rotation
	return &ClaudeProvider{cfg: cfg}
}

func (p *ClaudeProvider) Name() string {
	return string(ProviderClaude)
}

func (p *ClaudeProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
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
		client := anthropic.NewClient(option.WithAPIKey(currentKey))
		p.client = client

		maxTokens := req.MaxTokens
		if maxTokens == 0 {
			maxTokens = 1000
		}

		modelName := p.cfg.Model
		if req.Model != "" {
			modelName = req.Model
		}
		// Convert messages to Claude format
		var messages []anthropic.MessageParam
		if len(req.Messages) > 0 {
			messages = make([]anthropic.MessageParam, len(req.Messages))
			for i, msg := range req.Messages {
				role, _ := msg["role"].(string)
				content, _ := msg["content"].(string)

				switch role {
				case "user":
					messages[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(content))
				case "assistant":
					messages[i] = anthropic.NewAssistantMessage(anthropic.NewTextBlock(content))
				default:
					// Default to user message
					messages[i] = anthropic.NewUserMessage(anthropic.NewTextBlock(content))
				}
			}
		} else {
			// Fallback to single message
			messages = []anthropic.MessageParam{
				anthropic.NewUserMessage(anthropic.NewTextBlock(req.Prompt)),
			}
		}

		claudeReq := anthropic.MessageNewParams{
			Model:     anthropic.Model(modelName),
			MaxTokens: int64(maxTokens),
			Messages:  messages,
		}

		// Add params
		if temp, ok := req.Params["temperature"].(float64); ok {
			claudeReq.Temperature = anthropic.Float(temp)
		}
		if topP, ok := req.Params["top_p"].(float64); ok {
			claudeReq.TopP = anthropic.Float(topP)
		}

		resp, err := p.client.Messages.New(ctx, claudeReq)

		if err == nil && resp != nil && len(resp.Content) > 0 {
			var text string
			for _, block := range resp.Content {
				switch content := block.AsAny().(type) {
				case anthropic.TextBlock:
					text += content.Text
				}
			}

			if text != "" {
				tokensUsed := int(resp.Usage.InputTokens + resp.Usage.OutputTokens)
				// Update usage
				p.cfg.UpdateUsage(1, tokensUsed)

				return &LLMResponse{
					Text:         text,
					InputTokens:  int(resp.Usage.InputTokens),
					OutputTokens: int(resp.Usage.OutputTokens),
					TokensUsed:   tokensUsed,
					FinishReason: string(resp.StopReason),
				}, nil
			}
		}

		// If error and not last attempt, try next key
		if attempt < maxRetries-1 {
			p.cfg.NextAPIKey()
		} else {
			// Last attempt failed
			if err != nil {
				return nil, fmt.Errorf("Claude API error after %d attempts: %w", maxRetries, err)
			}
			return nil, fmt.Errorf("no response from Claude after %d attempts", maxRetries)
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

func (p *ClaudeProvider) GenerateStream(ctx context.Context, req *LLMRequest) (<-chan *LLMStreamResponse, error) {
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

func (p *ClaudeProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"claude-3-opus", "claude-3-sonnet", "claude-3-haiku"}, nil
}
