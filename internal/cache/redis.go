package cache

import (
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(redisURL string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       0,
		Protocol: 2,
	})
	return rdb
}
