package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/eviltwin7648/llm-gateway/internal/embedder"
	"github.com/redis/go-redis/v9"
)

type CacheEntry struct {
	PromptEmbedding []float32
	Prompt          string
	Response        string
	Model           string
	CreatedAt       time.Time
}
type Cahce interface {
	LookUp(ctx context.Context, embedding []float32) (*CacheEntry, bool)
	Store(ctx context.Context, entry CacheEntry) error
}

type Cache struct {
	rdb *redis.Client
}

func NewCache() *Cache {
	return &Cache{
		rdb: NewRedisClient(),
	}
}

func (c *Cache) Store(ctx context.Context, key string, value CacheEntry) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	err = c.rdb.Set(ctx, key, data, 0).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) LookUp(ctx context.Context, embedding embedder.Embedding) (*CacheEntry, error) {
	//senamtic similarity should be checked before returning, only return if score > 90% else return nil
	//normal redis does not support vector search so we either use redis stack or compute similarity myself

	data, err := c.rdb.Get(ctx, embedding).Result()
	if err != nil {
		return nil, err
	}
	var val Value
	if err := json.Unmarshal([]byte(data), &val); err != nil {
		return nil, err
	}
	return &val, nil
}
