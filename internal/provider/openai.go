package provider

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	cfg    LLMConfig
	client *openai.Client
}

func NewOpenAIProvider(cfg LLMConfig) *OpenAIProvider {
	config := openai.DefaultConfig(cfg.APIKey())
	if cfg.BaseURL != "" {
		config.BaseURL = cfg.BaseURL
	}
	client := openai.NewClientWithConfig(config)
	return &OpenAIProvider{cfg: cfg, client: client}
}

func (p *OpenAIProvider) Name() string {
	return string(ProviderOpenAI)
}

func (p *OpenAIProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
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
		if p.cfg.BaseURL != "" {
			config.BaseURL = p.cfg.BaseURL
		}
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
				return nil, fmt.Errorf("OpenAI API error after %d attempts: %w", maxRetries, err)
			}
			return nil, fmt.Errorf("no response from OpenAI after %d attempts", maxRetries)
		}
	}

	return nil, fmt.Errorf("unexpected error in retry loop")
}

func (p *OpenAIProvider) ListModels(ctx context.Context) ([]string, error) {
	// Implement list models if needed
	return []string{"gpt-4o", "gpt-4", "gpt-3.5-turbo"}, nil
}
