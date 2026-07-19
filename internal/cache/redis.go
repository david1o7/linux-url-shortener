package cache

import (
	"Linux-url-shortener/internal/logger"
	"context"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache() *RedisCache {
	
	var err = godotenv.Load()

	if err != nil{
	logger.Log.Error(
		".env file error",
		"Error", err,
	)
	}

	RedisAddr := os.Getenv("REDIS_ADDR")

	client := redis.NewClient(&redis.Options{
		Addr: "localhost"+RedisAddr,
		DB: 0,
	})

	return &RedisCache{
		Client: client,
	}
}

func (r *RedisCache) Set(key, value string) error {

	return r.Client.Set(
		ctx,
		key,
		value,
		24*60*60*1e9,
	).Err()

}

func (r *RedisCache) Get(key string) (string , error){
	return r.Client.Get(ctx, key).Result()
}