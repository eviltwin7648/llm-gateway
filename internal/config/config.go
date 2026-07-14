package config

import (
	"os"
)

type AppConfig struct {
	Port              string
	RedisURL          string
	OpenAIAPIKey      string
	GeminiAPIKey      string
	AnthropicAPIKey   string
	DeepSeekAPIKey    string
	OpenAIBaseURL     string
	DeepSeekBaseURL   string
}

func Load() *AppConfig {
	return &AppConfig{
		Port:            getEnvOrDefault("PORT", "8080"),
		RedisURL:        getEnvOrDefault("REDIS_URL", "localhost:6379"),
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		DeepSeekAPIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		OpenAIBaseURL:   os.Getenv("OPENAI_BASE_URL"),
		DeepSeekBaseURL: getEnvOrDefault("DEEPSEEK_BASE_URL", "https://api.deepseek.com"),
	}
}

func getEnvOrDefault(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
