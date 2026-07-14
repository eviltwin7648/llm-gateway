package cache

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CacheEntry struct {
	Embeddings []float32
	Prompt     string
	Response   string
	Provider   string
	Model      string
	Temprature float32
	HitCount   int
	LastUsed   time.Time
	CreatedAt  time.Time
}
type LookupFilter struct {
	Provider string
	Model    string
}
type CacheStore interface {
	LookUp(ctx context.Context, embedding []float32, filter LookupFilter) (*CacheEntry, bool, error)
	Store(ctx context.Context, entry CacheEntry) error
}

type RedisCache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *RedisCache {
	return &RedisCache{
		rdb: rdb,
	}
}

func CreateIndex(ctx context.Context, rdb *redis.Client) error {
	_, err := rdb.Do(ctx,
		"FT.CREATE",
		"idx:embeddings",
		"ON", "HASH",
		"PREFIX", "1", "cache:",
		"SCHEMA",
		"embedding",
		"VECTOR", "HNSW", "6",
		"TYPE", "FLOAT32",
		"DIM", "1536",
		"DISTANCE_METRIC", "COSINE",

		"provider", "TAG",
		"model", "TAG",

		"temperature", "NUMERIC",
		"created_at", "NUMERIC",
		"last_used", "NUMERIC",
		"hit_count", "NUMERIC",
	).Result()
	if err != nil {
		if strings.Contains(err.Error(), "Index already exists") {
			return nil
		}
		return err
	}
	return nil
}

func (c *RedisCache) Store(ctx context.Context, entry CacheEntry) error {
	//we need to store the embeddings in a hash, create a HNSW index for effective vector search
	// raw json cannot be stored and indexed as it is. it should be converted into a binary blob.

	if len(entry.Embeddings) != 1536 {
		return fmt.Errorf("invalid embedding dimension: got %d, want 1536", len(entry.Embeddings))
	}
	if entry.Prompt == "" || entry.Response == "" {
		return fmt.Errorf("prompt and response cannot be empty")
	}
	now := time.Now()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.LastUsed = now

	key := "cache:" + uuid.NewString()
	embeddingBlob := float32SliceToBinary(entry.Embeddings)

	err := c.rdb.HSet(ctx, key, map[string]any{
		"embedding":   embeddingBlob,
		"prompt":      entry.Prompt,
		"response":    entry.Response,
		"provider":    entry.Provider,
		"model":       entry.Model,
		"temperature": entry.Temprature,
		"hit_count":   entry.HitCount,
		"last_used":   entry.LastUsed.UnixMilli(),
		"created_at":  entry.CreatedAt.UnixMilli(),
	}).Err()
	if err != nil {
		return err
	}
	return nil
}

func (c *RedisCache) LookUp(ctx context.Context, embedding []float32, filter LookupFilter) (*CacheEntry, bool, error) {
	//senamtic similarity should be checked before returning, only return if score > 90% else return nil
	//lookup should have some metadata filtering, along with vector search
	//hybrid search : metadata filter  + vector search
	if len(embedding) != 1536 {
		return nil, false, fmt.Errorf("invalid embedding dimension: got %d, want 1536", len(embedding))
	}

	var filters []string
	if filter.Provider != "" {
		filters = append(filters, fmt.Sprintf("@provider:{%s}", escapeTag(filter.Provider)))
	}
	if filter.Model != "" {
		filters = append(filters, fmt.Sprintf("@model:{%s}", escapeTag(filter.Model)))
	}

	filterQuery := "*"
	if len(filters) > 0 {
		filterQuery = "(" + strings.Join(filters, " ") + ")"
	}

	query := fmt.Sprintf(
		"%s=>[KNN 1 @embedding $query_vec AS distance]",
		filterQuery,
	)

	queryblob := float32SliceToBinary(embedding)
	results, err := c.rdb.Do(
		ctx,
		"FT.SEARCH",
		"idx:embeddings",
		query,
		"PARAMS", "2",
		"query_vec", queryblob,
		"SORTBY", "distance",
		"RETURN", "9", "prompt", "response", "provider", "model", "temperature", "hit_count", "last_used", "created_at", "distance",
		//wihtout this it will only return First found value (https://redis.io/docs/latest/commands/ft.search/)

		"DIALECT", "2",
	).Result()
	if err != nil {
		return nil, false, fmt.Errorf("search semantic cache: %w", err)
	}
	// result is mostly flat array from redis,so we tell go we expect an array of generic interface
	resList, ok := results.([]interface{})
	if !ok || len(resList) == 0 {
		return nil, false, fmt.Errorf("unexpected results format from FT.SEARCH")
	}
	//first element is the total number of results
	totalResults, ok := resList[0].(int64)
	if !ok {
		return nil, false, fmt.Errorf("unexpected total results format")
	}
	if totalResults == 0 {
		return nil, false, nil // Cache miss
	}

	// succeeful response with atleast one value will have array of 3 elements
	// [totalResults, keys([]interface{}), values([]interface{})]
	if len(resList) < 3 {
		return nil, false, fmt.Errorf("unexpected results length from FT.SEARCH")
	}

	fields, ok := resList[2].([]interface{})
	if !ok {
		return nil, false, fmt.Errorf("unexpected fields format from FT.SEARCH")
	}

	entry := &CacheEntry{}
	var distance float64
	// map the returned data to the cacheEntry struct fields
	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		valStr, isStr := fields[i+1].(string)
		if !isStr {
			continue
		}

		switch key {
		case "prompt":
			entry.Prompt = valStr
		case "response":
			entry.Response = valStr
		case "provider":
			entry.Provider = valStr
		case "model":
			entry.Model = valStr
		case "temperature":
			if f, err := strconv.ParseFloat(valStr, 32); err == nil {
				entry.Temprature = float32(f)
			}
		case "hit_count":
			if d, err := strconv.Atoi(valStr); err == nil {
				entry.HitCount = d
			}
		case "last_used":
			if d, err := strconv.ParseInt(valStr, 10, 64); err == nil {
				entry.LastUsed = time.UnixMilli(d)
			}
		case "created_at":
			if d, err := strconv.ParseInt(valStr, 10, 64); err == nil {
				entry.CreatedAt = time.UnixMilli(d)
			}
		case "distance":
			if f, err := strconv.ParseFloat(valStr, 64); err == nil {
				distance = f
			}
		}
	}

	// semantic similarity > 90% implies cosine distance <= 0.1
	if distance > 0.1 {
		return nil, false, nil // Cache miss, not similar enough
	}

	return entry, true, nil
}

// function to convert float32  to binary
func float32SliceToBinary(v []float32) []byte {
	//since 8 bits = 1 byte
	//32 bits = 4bytes
	//so we need to create a buffer of size len(v)*4

	//for fucks sake its a little confusing, need to brush up basics.
	buf := make([]byte, len(v)*4)
	for i, f := range v {
		binary.LittleEndian.PutUint32(
			buf[i*4:],
			//converts float32 to binary bit
			math.Float32bits(f),
		)
	}
	return buf
}

func escapeTag(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`,`, `\,`,
		`{`, `\{`,
		`}`, `\}`,
		`|`, `\|`,
	)

	return replacer.Replace(s)
}
