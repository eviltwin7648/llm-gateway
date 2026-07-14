package service

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/eviltwin7648/llm-gateway/internal/cache"
	"github.com/eviltwin7648/llm-gateway/internal/embedder"
	"github.com/eviltwin7648/llm-gateway/internal/model"
	"github.com/eviltwin7648/llm-gateway/internal/router"
	"github.com/eviltwin7648/llm-gateway/internal/usage"
)

type GatewayService struct {
	cache    cache.CacheStore
	embedder embedder.Embedder
	router   *router.Router
	usage    usage.Recorder
}

func NewGatewayService(c cache.CacheStore, e embedder.Embedder, r *router.Router, u usage.Recorder) *GatewayService {
	return &GatewayService{
		cache:    c,
		embedder: e,
		router:   r,
		usage:    u,
	}
}

func (g *GatewayService) HandleChat(ctx context.Context, req model.ChatRequest) (model.ChatResponse, error) {
	//normalize req
	normalizedReq := g.normalize(req)
	//embedd req
	embedding, err := g.embedder.Embed(ctx, normalizedReq.Prompt)
	if err != nil {
		return model.ChatResponse{}, err
	}
	//check cache
	hit, found, err := g.cache.LookUp(ctx, embedding.Values, cache.LookupFilter{
		Provider: normalizedReq.Provider,
		Model:    normalizedReq.Model,
	})
	if err != nil {
		return model.ChatResponse{}, err
	}
	//return if found
	if found && hit != nil {
		return model.ChatResponse{
			Content: hit.Response,
			Model:   hit.Model,
			//since its a cached req, i'm setting the usage and token as 0
			Usage:  0,
			Tokens: 0,
		}, nil
	}

	// if not found || route based on model name
	//router should return the provider implementation not the enum
	providerImpl, err := g.router.Route(normalizedReq.Model)
	if err != nil {
		return model.ChatResponse{}, err
	}
	resp, err := providerImpl.Chat(ctx, normalizedReq)
	if err != nil {
		return model.ChatResponse{}, err
	}
	//cache response
	if err := g.cache.Store(ctx, cache.CacheEntry{
		Embeddings: embedding.Values,
		Prompt:     normalizedReq.Prompt,
		Response:   resp.Content,
		Provider:   normalizedReq.Provider,
		Model:      normalizedReq.Model,

		// Temprature: normalizedReq.Temprature,
		HitCount:  0,
		LastUsed:  time.Now(),
		CreatedAt: time.Now(),
	}); err != nil {
		log.Printf("Failed to store response in cache: %v", err)
	}
	//record usage
	if err := g.usage.Record(ctx, normalizedReq, *resp); err != nil {
		log.Printf("Failed to record usage: %v", err)
	}
	//return response
	return *resp, nil
}

func (g *GatewayService) normalize(req model.ChatRequest) model.ChatRequest {
	providerImpl, err := g.router.Route(req.Model)
	if err != nil {
		return model.ChatRequest{}
	}
	return model.ChatRequest{
		Prompt:   strings.TrimSpace(req.Prompt),
		Model:    strings.ToLower(strings.TrimSpace(req.Model)),
		Provider: providerImpl.Name(),
	}
}
