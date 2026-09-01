package cache

import (
	"context"
	"fmt"
	"time"

	"Linux-url-shortener/internal/config"
	"Linux-url-shortener/internal/logger"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache(cfg *config.Config) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           cfg.RedisDB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	return &RedisCache{Client: client}
}

func (r *RedisCache) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

func (r *RedisCache) Set(key, value string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return r.Client.Set(ctx, key, value, 24*time.Hour).Err()
}

func (r *RedisCache) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	val, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found")
	}
	return val, err
}

func (r *RedisCache) Close() error {
	if r.Client == nil {
		return nil
	}
	err := r.Client.Close()
	if err != nil {
		logger.Log.Error("Redis close failed", "error", err)
		return err
	}
	logger.Log.Info("Redis connection closed")
	return nil
}
