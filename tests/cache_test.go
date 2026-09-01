package tests

import (
	"Linux-url-shortener/internal/cache"
	"testing"
)

func TestCache_NewCache(t *testing.T) {
	c := cache.NewCache()

	if c == nil {
		t.Fatalf("expected cache, got nil")
	}
}

func TestCache_GetMissingKey(t *testing.T) {
	c := cache.NewCache()

	value, ok := c.Get("missing")

	if ok {
		t.Fatalf("Expected Missing Key to return False")
	}

	if value != "" {
		t.Fatalf("expected empty value, got %q", value)
	}
}

func TestCache_SetAndGet(t *testing.T) {
	c := cache.NewCache()

	c.Set("abc123", "https://google.com")

	value, ok := c.Get("abc123")

	if !ok {
		t.Fatal("Expected key to exist")
	}

	if value != "https://google.com" {
		t.Fatalf("expected https://google.com, got %q", value)
	}
}

func TestCache_SetOverWriteExistingValue(t *testing.T) {
	c := cache.NewCache()

	c.Set("abc123", "https://google.com")
	c.Set("abc123", "https://github.com")

	value, ok := c.Get("abc123")

	if !ok {
		t.Fatalf("Expected key to exist")
	}

	if value != "https://github.com" {
		t.Fatalf("Expected updated value, got %q", value)
	}
}
