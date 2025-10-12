package provider

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeProvider struct {
	cfg    LLMConfig
	client anthropic.Client
}

func NewClaudeProvider(cfg LLMConfig) *ClaudeProvider {
	client := anthropic.NewClient(option.WithAPIKey(cfg.APIKey()))
	return &ClaudeProvider{cfg: cfg, client: client}
}

func (p *ClaudeProvider) Name() string {
	return string(ProviderClaude)
}

func (p *ClaudeProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
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
	if err != nil {
		return nil, fmt.Errorf("Claude API error: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("nil response from Claude API")
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("no response from Claude")
	}

	var text string
	for _, block := range resp.Content {
		switch content := block.AsAny().(type) {
		case anthropic.TextBlock:
			text += content.Text
		}
	}

	if text == "" {
		return nil, fmt.Errorf("no text content from Claude")
	}

	return &LLMResponse{
		Text:         text,
		InputTokens:  int(resp.Usage.InputTokens),
		OutputTokens: int(resp.Usage.OutputTokens),
		TokensUsed:   int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		FinishReason: string(resp.StopReason),
	}, nil
}

func (p *ClaudeProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"claude-3-opus-20240229", "claude-3-sonnet-20240229"}, nil
}
