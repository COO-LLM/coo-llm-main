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

func (p *GrokProvider) ListModels(ctx context.Context) ([]string, error) {
	// Grok typically supports grok-beta and similar models
	return []string{"grok-beta"}, nil
}
