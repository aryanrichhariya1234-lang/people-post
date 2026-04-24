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

func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic(err)
	}

	if opt.TLSConfig == nil {
		opt.TLSConfig = &tls.Config{}
	}

	Client = redis.NewClient(opt)

	_, err = Client.Ping(ctx).Result()
	if err != nil {
		panic("Redis not connected: " + err.Error())
	}
}

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

func Get(key string, dest interface{}) error {
	if Client == nil {
		return errors.New("redis not initialized")
	}

	val, err := Client.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func Delete(key string) error {
	if Client == nil {
		return errors.New("redis not initialized")
	}

	return Client.Del(ctx, key).Err()
}