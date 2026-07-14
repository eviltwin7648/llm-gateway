// just for console logging
package usage

import (
	"context"
	"log"

	"github.com/eviltwin7648/llm-gateway/internal/model"
)

// create implementations for whatever service u need to use (sql , redis etc)
type Recorder interface {
	Record(ctx context.Context, req model.ChatRequest, resp model.ChatResponse) error
}

type LogRecorder struct{}

func NewLogRecorder() *LogRecorder {
	return &LogRecorder{}
}

func (u *LogRecorder) Record(ctx context.Context, req model.ChatRequest, resp model.ChatResponse) error {
	log.Printf("[USAGE] Provider: %s | Model: %s | Tokens: %d | UsageCost: %d", req.Provider, resp.Model, resp.Tokens, resp.Usage)
	return nil
}
