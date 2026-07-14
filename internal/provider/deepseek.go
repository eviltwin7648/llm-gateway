package provider

//the deepseek api is designed to be fully compatioble with openai api format. so we use the openai client and overwrite the baseurl
import (
	"context"

	"github.com/eviltwin7648/llm-gateway/internal/model"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type DeepSeekProvider struct {
	client openai.Client
}

func NewDeepSeekProvider(apiKey string, baseURL string) *DeepSeekProvider {
	opts := []option.RequestOption{option.WithAPIKey(apiKey)}
	if baseURL != "" {
		//base url of deepseek get overwritten
		opts = append(opts, option.WithBaseURL(baseURL))
	}
	return &DeepSeekProvider{
		client: openai.NewClient(opts...),
	}
}

func (p *DeepSeekProvider) Chat(ctx context.Context, req model.ChatRequest) (*model.ChatResponse, error) {
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

func (p *DeepSeekProvider) Name() string {
	return "deepseek"
}
