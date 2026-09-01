package tests

import (
	"Linux-url-shortener/internal/cache"
	"Linux-url-shortener/internal/config"
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisCache_SetGetPingClose(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	cfg := &config.Config{RedisAddr: mr.Addr()}
	rc := cache.NewRedisCache(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rc.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}

	if err := rc.Set("k1", "https://example.com"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	val, err := rc.Get("k1")
	if err != nil || val != "https://example.com" {
		t.Fatalf("Get = %q, %v", val, err)
	}

	_, err = rc.Get("missing")
	if err == nil {
		t.Fatal("expected error for missing key")
	}

	if err := rc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestRedisCache_CloseNilClient(t *testing.T) {
	rc := &cache.RedisCache{Client: nil}
	if err := rc.Close(); err != nil {
		t.Fatalf("Close nil: %v", err)
	}
}

func TestRedisCache_GetWhenDown(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	addr := mr.Addr()
	mr.Close()

	client := redis.NewClient(&redis.Options{Addr: addr})
	rc := &cache.RedisCache{Client: client}
	_, err = rc.Get("x")
	if err == nil {
		t.Fatal("expected error when redis is down")
	}
}
