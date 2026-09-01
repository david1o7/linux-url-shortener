package integration

import (
	"context"
	"testing"
	"time"
)

func TestIntegration_RedisCacheRoundTrip(t *testing.T) {
	resetState(t)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := shared.Cache.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}

	if err := shared.Cache.Set("int-k", "https://cached.example"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	val, err := shared.Cache.Get("int-k")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if val != "https://cached.example" {
		t.Fatalf("val=%q", val)
	}

	_, err = shared.Cache.Get("does-not-exist")
	if err == nil {
		t.Fatal("expected miss error")
	}
}
