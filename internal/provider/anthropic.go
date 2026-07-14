package provider

import (
	"context"

	"github.com/eviltwin7648/llm-gateway/internal/model"
	"github.com/liushuangls/go-anthropic/v2"
)

type AnthropicProvider struct {
	client *anthropic.Client
}

func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{
		client: anthropic.NewClient(apiKey),
	}
}

func (p *AnthropicProvider) Chat(ctx context.Context, req model.ChatRequest) (*model.ChatResponse, error) {
	resp, err := p.client.CreateMessages(ctx, anthropic.MessagesRequest{
		Model: anthropic.Model(req.Model),
		Messages: []anthropic.Message{
			{
				Role: anthropic.RoleUser,
				Content: []anthropic.MessageContent{
					anthropic.NewTextMessageContent(req.Prompt),
				},
			},
		},
		MaxTokens: 4096,
	})

	if err != nil {
		return nil, err
	}

	content := ""
	for _, c := range resp.Content {
		if c.Type == anthropic.MessagesContentTypeText {
			content += *c.Text
		}
	}

	tokens := resp.Usage.InputTokens + resp.Usage.OutputTokens

	return &model.ChatResponse{
		Content: content,
		Usage:   tokens,
		Tokens:  tokens,
		Model:   req.Model,
	}, nil
}

func (p *AnthropicProvider) Name() string {
	return "anthropic"
}
