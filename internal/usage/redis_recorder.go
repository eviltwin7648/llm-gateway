// loggin in redis
package usage

import (
	"context"

	"github.com/eviltwin7648/llm-gateway/internal/model"
	"github.com/redis/go-redis/v9"
)

type RedisRecorder struct {
	rdb *redis.Client
}

func NewRedisRecorder(rdb *redis.Client) *RedisRecorder {
	return &RedisRecorder{
		rdb: rdb,
	}
}

func (r *RedisRecorder) Record(ctx context.Context, req model.ChatRequest, resp model.ChatResponse) error {
	// We use Redis Streams to store the time-series usage logs
	err := r.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "usage_metrics",
		Values: map[string]interface{}{
			"provider": req.Provider,
			"model":    resp.Model,
			"tokens":   resp.Tokens,
			"usage":    resp.Usage,
			"prompt":   req.Prompt,
		},
	}).Err()

	return err
}
