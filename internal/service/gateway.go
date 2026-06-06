package service

import (
	"context"

	"github.com/eviltwin7648/llm-gateway/internal/cache"
	"github.com/eviltwin7648/llm-gateway/internal/embedder"
	"github.com/eviltwin7648/llm-gateway/internal/model"
	"github.com/eviltwin7648/llm-gateway/internal/router"
	"github.com/eviltwin7648/llm-gateway/internal/usage"
)

type GatewayService struct {
	cache    cache.Cache
	embedder embedder.Embedder
	router   router.Router
	usage    usage.UsageRecorder
}

func (g *GatewayService) HandleChat(ctx context.Context, req model.ChatRequest) model.ChatResponse {
	embedding := g.embedder.Embed(ctx, req.Prompt)

	hit := g.cache.GetData(ctx, embedding)

	if hit {
		return hit.Response //type mismatch
	}

	provider := g.router.Route(req.Model)

	resp := provider.Chat(req)

	g.usage.Record() // pass whatever required
	g.cache.SetData(resp)
	return resp

}
