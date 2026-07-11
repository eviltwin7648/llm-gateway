package router

import (
	"fmt"

	"github.com/eviltwin7648/llm-gateway/internal/provider"
)

type Router struct {
	models map[string]provider.Provider
}

func NewRouter(openAI, google, anthropic, deepseek provider.Provider) *Router {
	r := &Router{
		models: make(map[string]provider.Provider),
	}

	//registering all models
	if openAI != nil {
		r.models["gpt-5"] = openAI
		r.models["gpt-5-mini"] = openAI
		r.models["gpt-5-nano"] = openAI
		r.models["gpt-5.1"] = openAI
		r.models["gpt-5-pro"] = openAI

		r.models["gpt-4.1"] = openAI
		r.models["gpt-4.1-mini"] = openAI
		r.models["gpt-4o"] = openAI
		r.models["gpt-4o-mini"] = openAI

		r.models["o3"] = openAI
		r.models["o3-pro"] = openAI
		r.models["o4-mini"] = openAI

		r.models["gpt-oss-120b"] = openAI
		r.models["gpt-oss-20b"] = openAI
	}

	// Google
	if google != nil {
		r.models["gemini-2.5-pro"] = google
		r.models["gemini-2.5-flash"] = google
		r.models["gemini-2.5-flash-lite"] = google
	}

	// Anthropic
	if anthropic != nil {
		r.models["claude-opus-4"] = anthropic
		r.models["claude-opus-4.5"] = anthropic

		r.models["claude-sonnet-4"] = anthropic
		r.models["claude-sonnet-4.5"] = anthropic

		r.models["claude-haiku-4.5"] = anthropic
	}

	// DeepSeek
	if deepseek != nil {
		r.models["deepseek-chat"] = deepseek
		r.models["deepseek-reasoner"] = deepseek
	}
	return r
}

func (r *Router) Route(model string) (provider.Provider, error) {
	provider, ok := r.models[model]
	if !ok {
		return nil, fmt.Errorf("unknown model: %s", model)
	}

	return provider, nil
}
