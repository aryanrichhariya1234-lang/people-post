package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()
var Client *redis.Client

// Initialize Redis
func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		panic("REDIS_URL not set")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}

	// Ensure TLS (required for many hosted Redis providers)
	if opt.TLSConfig == nil {
		opt.TLSConfig = &tls.Config{}
	}

	Client = redis.NewClient(opt)

	_, err = Client.Ping(ctx).Result()
	if err != nil {
		panic("Redis not connected: " + err.Error())
	}
}

// Set value with expiration (TTL)
func Set(key string, value interface{}, expiration time.Duration) error {
	if Client == nil {
		return errors.New("redis not initialized")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return Client.Set(ctx, key, data, expiration).Err()
}

// Get value from cache
func Get(key string, dest interface{}) error {
	if Client == nil {
		return errors.New("redis not initialized")
	}

	val, err := Client.Get(ctx, key).Result()

	// Handle cache miss properly
	if err == redis.Nil {
		return errors.New("cache miss")
	}
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// Delete a key manually (useful for invalidation)
func Delete(key string) error {
	if Client == nil {
		return errors.New("redis not initialized")
	}

	return Client.Del(ctx, key).Err()
}

// Optional: helper to check if key exists
func Exists(key string) (bool, error) {
	if Client == nil {
		return false, errors.New("redis not initialized")
	}

	count, err := Client.Exists(ctx, key).Result()
	return count > 0, err
}

// Optional: helper to set only if key does not exist
func SetNX(key string, value interface{}, expiration time.Duration) error {
	if Client == nil {
		return errors.New("redis not initialized")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return Client.SetNX(ctx, key, data, expiration).Err()
}