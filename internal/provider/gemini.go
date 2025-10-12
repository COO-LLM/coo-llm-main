package provider

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type GeminiProvider struct {
	cfg    LLMConfig
	client *genai.Client
}

func NewGeminiProvider(cfg LLMConfig) *GeminiProvider {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey()))
	if err != nil {
		// Handle error, but for now panic or return nil
		panic(fmt.Sprintf("failed to create Gemini client: %v", err))
	}
	return &GeminiProvider{cfg: cfg, client: client}
}

func (p *GeminiProvider) Name() string {
	return string(ProviderGemini)
}

func (p *GeminiProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
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
	var err error
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
	if err != nil {
		return nil, fmt.Errorf("Gemini API error: %w", err)
	}

	if resp == nil {
		return nil, fmt.Errorf("nil response from Gemini API")
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	text, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return nil, fmt.Errorf("unexpected response type from Gemini")
	}

	// Estimate tokens (SDK may not provide exact count)
	inputTokens := len(req.Prompt) / 4
	outputTokens := len(string(text)) / 4
	tokensUsed := inputTokens + outputTokens

	return &LLMResponse{
		Text:         string(text),
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TokensUsed:   tokensUsed,
		FinishReason: string(resp.Candidates[0].FinishReason),
	}, nil
}

func (p *GeminiProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"gemini-1.5-pro", "gemini-1.5-flash"}, nil
}
