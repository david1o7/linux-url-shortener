package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache() *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
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