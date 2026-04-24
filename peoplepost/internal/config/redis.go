package config

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("UPSTASH_REDIS_REST_URL"),
		Password: os.Getenv("UPSTASH_REDIS_REST_TOKEN"),
		DB:       0,
	})
}