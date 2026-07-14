package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/eviltwin7648/llm-gateway/internal/cache"
	"github.com/eviltwin7648/llm-gateway/internal/config"
	"github.com/eviltwin7648/llm-gateway/internal/embedder"
	"github.com/eviltwin7648/llm-gateway/internal/model"
	"github.com/eviltwin7648/llm-gateway/internal/provider"
	"github.com/eviltwin7648/llm-gateway/internal/router"
	"github.com/eviltwin7648/llm-gateway/internal/service"
	"github.com/eviltwin7648/llm-gateway/internal/usage"
)

func main() {
	// Load Configuration
	cfg := config.Load()
	log.Printf("Starting LLM Gateway on port %s", cfg.Port)

	// Initialize Providers
	var openAIProvider, geminiProvider, anthropicProvider, deepSeekProvider provider.Provider

	if cfg.OpenAIAPIKey != "" {
		openAIProvider = provider.NewOpenAIProvider(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL)
	}
	if cfg.GeminiAPIKey != "" {
		geminiProvider = provider.NewGeminiProvider(cfg.GeminiAPIKey)
	}
	if cfg.AnthropicAPIKey != "" {
		anthropicProvider = provider.NewAnthropicProvider(cfg.AnthropicAPIKey)
	}
	if cfg.DeepSeekAPIKey != "" {
		//overwrite base url
		deepSeekProvider = provider.NewDeepSeekProvider(cfg.DeepSeekAPIKey, cfg.DeepSeekBaseURL)
	}

	// Initialize Router
	r := router.NewRouter(openAIProvider, geminiProvider, anthropicProvider, deepSeekProvider)

	// Initialize Shared Redis Client
	redisClient := cache.NewRedisClient(cfg.RedisURL)

	// Initialize Cache, Embedder, and Usage Recorder
	c := cache.NewCache(redisClient)
	var e embedder.Embedder
	if cfg.OpenAIAPIKey != "" {
		e = embedder.NewEmbedder(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL)
	} else {
		log.Println("WARNING: Embedder is nil because OPENAI_API_KEY is not set.")
	}
	u := usage.NewRedisRecorder(redisClient)

	// Initialize Gateway Service
	gatewayService := service.NewGatewayService(c, e, r, u)

	// Setup HTTP Server
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var chatReq model.ChatRequest
		if err := json.NewDecoder(req.Body).Decode(&chatReq); err != nil {
			http.Error(w, "Invalid Request Body", http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		resp, err := gatewayService.HandleChat(req.Context(), chatReq)
		if err != nil {
			log.Printf("Error handling chat: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	})

	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
