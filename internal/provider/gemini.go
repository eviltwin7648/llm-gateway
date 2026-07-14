package provider

import (
	"context"
	"fmt"

	"github.com/eviltwin7648/llm-gateway/internal/model"
	"google.golang.org/genai"
)

type GeminiProvider struct {
	apiKey string
}

func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{
		apiKey: apiKey,
	}
}

func (p *GeminiProvider) Chat(ctx context.Context, req model.ChatRequest) (*model.ChatResponse, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: p.apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	resp, err := client.Models.GenerateContent(ctx, req.Model, genai.Text(req.Prompt), nil)
	if err != nil {
		return nil, err
	}

	content := ""
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				content += part.Text
			}
		}
	}

	tokens := 0
	if resp.UsageMetadata != nil {
		tokens = int(resp.UsageMetadata.TotalTokenCount)
	}

	return &model.ChatResponse{
		Content: content,
		Usage:   tokens,
		Tokens:  tokens,
		Model:   req.Model,
	}, nil
}

func (p *GeminiProvider) Name() string {
	return "google"
}
