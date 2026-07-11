package provider

import (
	"context"

	"github.com/eviltwin7648/llm-gateway/internal/model"
)

type Provider interface {
	Chat(ctx context.Context, req model.ChatRequest) (*model.ChatResponse, error)
	Name() string
}
