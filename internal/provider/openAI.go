package provider

import (
	"context"

	"github.com/eviltwin7648/llm-gateway/internal/model"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAIProvider struct {
	client openai.Client
}

func NewOpenAIProvider(apiKey string, baseURL string) *OpenAIProvider {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return &OpenAIProvider{
		client: openai.NewClient(opts...),
	}
}

func (p *OpenAIProvider) Chat(ctx context.Context, req model.ChatRequest) (*model.ChatResponse, error) {
	completion, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModel(req.Model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(req.Prompt),
		},
	})
	if err != nil {
		return nil, err
	}

	content := completion.Choices[0].Message.Content
	tokens := int(completion.Usage.TotalTokens)

	return &model.ChatResponse{
		Content: content,
		Usage:   tokens,
		Tokens:  tokens,
		Model:   req.Model,
	}, nil
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}
