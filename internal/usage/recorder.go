package usage

import (
	"log"

	"github.com/eviltwin7648/llm-gateway/internal/model"
)

type UsageRecorder struct {
}

func NewUsageRecorder() *UsageRecorder {
	return &UsageRecorder{}
}

func (u *UsageRecorder) Record(req model.ChatRequest, resp model.ChatResponse) {
	// For now, we will simply log the usage to stdout.
	// In a real implementation, this would insert a row into a database (e.g. Postgres, ClickHouse).
	log.Printf("[USAGE] Provider: %s | Model: %s | Tokens: %d | UsageCost: %d", req.Provider, resp.Model, resp.Tokens, resp.Usage)
}
