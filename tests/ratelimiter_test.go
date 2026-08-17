package tests

import (
	"Linux-url-shortener/internal/middleware"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func setupRateLimiter(t *testing.T, capacity int64, refillRate float64) (*middleware.RateLimiter, *miniredis.Miniredis) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cfg := middleware.TokenBucketConfig{
		Capacity:   capacity,
		RefillRate: refillRate,
		KeyPrefix:  "tb:test:",
		TTL:        2 * time.Minute,
	}

	rl := middleware.NewRateLimiterWithConfig(client, cfg)
	return rl, mr
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func doRequest(rl *middleware.RateLimiter, ip string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/shorten", nil)
	req.RemoteAddr = ip + ":12345"
	rec := httptest.NewRecorder()
	rl.Limit(okHandler()).ServeHTTP(rec, req)
	return rec
}

func TestTokenBucket_AllowsUpToCapacity(t *testing.T) {
	rl, mr := setupRateLimiter(t, 5, 0.001)
	defer mr.Close()

	const ip = "10.0.0.1"

	for i := 0; i < 5; i++ {
		rec := doRequest(rl, ip)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}

func TestTokenBucket_RejectsWhenEmpty(t *testing.T) {
	rl, mr := setupRateLimiter(t, 3, 0.001)
	defer mr.Close()

	const ip = "10.0.0.2"

	// exhaust the bucket
	for i := 0; i < 3; i++ {
		if rec := doRequest(rl, ip); rec.Code != http.StatusOK {
			t.Fatalf("setup request %d failed: %d", i+1, rec.Code)
		}
	}

	// next one must be 429
	rec := doRequest(rl, ip)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}

	// Retry-After header should be present and positive
	ra := rec.Header().Get("Retry-After")
	if ra == "" {
		t.Fatal("expected Retry-After header")
	}
	n, err := strconv.Atoi(ra)
	if err != nil || n <= 0 {
		t.Fatalf("Retry-After should be a positive integer, got %q", ra)
	}
}

func TestTokenBucket_DifferentIPsAreIndependent(t *testing.T) {
	rl, mr := setupRateLimiter(t, 2, 0.001)
	defer mr.Close()

	// exhaust IP-A
	for i := 0; i < 2; i++ {
		if rec := doRequest(rl, "10.0.0.10"); rec.Code != http.StatusOK {
			t.Fatalf("IP-A request %d: %d", i+1, rec.Code)
		}
	}
	if rec := doRequest(rl, "10.0.0.10"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("IP-A should be limited, got %d", rec.Code)
	}

	// IP-B still has full capacity
	for i := 0; i < 2; i++ {
		if rec := doRequest(rl, "10.0.0.20"); rec.Code != http.StatusOK {
			t.Fatalf("IP-B request %d: %d", i+1, rec.Code)
		}
	}
}

func TestTokenBucket_SetsRateLimitHeaders(t *testing.T) {
	rl, mr := setupRateLimiter(t, 10, 1.0)
	defer mr.Close()

	rec := doRequest(rl, "10.0.0.3")

	if rec.Header().Get("X-RateLimit-Limit") != "10" {
		t.Errorf("X-RateLimit-Limit = %q, want 10", rec.Header().Get("X-RateLimit-Limit"))
	}

	remaining := rec.Header().Get("X-RateLimit-Remaining")
	if remaining == "" {
		t.Fatal("expected X-RateLimit-Remaining")
	}
	n, err := strconv.ParseInt(remaining, 10, 64)
	if err != nil {
		t.Fatalf("X-RateLimit-Remaining not an integer: %v", err)
	}
	// after one request we should have roughly capacity-1 left
	if n < 8 || n > 9 {
		t.Errorf("X-RateLimit-Remaining = %d, expected around 9", n)
	}
}

func TestTokenBucket_RefillsOverTime(t *testing.T) {
	// capacity 2, refill 10 tokens/sec → full refill in ~0.2 s
	rl, mr := setupRateLimiter(t, 2, 10.0)
	defer mr.Close()

	const ip = "10.0.0.4"

	// burn both tokens
	for i := 0; i < 2; i++ {
		if rec := doRequest(rl, ip); rec.Code != http.StatusOK {
			t.Fatalf("burn %d: %d", i+1, rec.Code)
		}
	}

	// should be limited
	if rec := doRequest(rl, ip); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 right after burn, got %d", rec.Code)
	}

	// wait long enough for at least one token to refill
	time.Sleep(150 * time.Millisecond)

	rec := doRequest(rl, ip)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 after refill, got %d", rec.Code)
	}
}

func TestTokenBucket_ConcurrentRequestsDoNotExceedCapacity(t *testing.T) {
	const capacity = 20
	rl, mr := setupRateLimiter(t, capacity, 0.0001) // virtually no refill
	defer mr.Close()

	const ip = "10.0.0.5"
	const goroutines = 50

	var allowed int64
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			rec := doRequest(rl, ip)
			if rec.Code == http.StatusOK {
				atomic.AddInt64(&allowed, 1)
			}
		}()
	}
	wg.Wait()

	if allowed > capacity {
		t.Fatalf("allowed %d requests, capacity is only %d - race condition!", allowed, capacity)
	}
	// under extreme scheduling we might get a few less, but never more
	if allowed < capacity-2 {
		t.Logf("warning: only %d of %d tokens used (possible under high contention)", allowed, capacity)
	}
}

func TestTokenBucket_UsesXForwardedFor(t *testing.T) {
	rl, mr := setupRateLimiter(t, 1, 0.001)
	defer mr.Close()

	// first request with XFF
	req1 := httptest.NewRequest(http.MethodGet, "/", nil)
	req1.RemoteAddr = "127.0.0.1:9999"
	req1.Header.Set("X-Forwarded-For", "203.0.113.10")
	rec1 := httptest.NewRecorder()
	rl.Limit(okHandler()).ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusOK {
		t.Fatalf("first XFF request: %d", rec1.Code)
	}

	// second request same XFF → should be limited
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "127.0.0.1:9999"
	req2.Header.Set("X-Forwarded-For", "203.0.113.10")
	rec2 := httptest.NewRecorder()
	rl.Limit(okHandler()).ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("second XFF request should be 429, got %d", rec2.Code)
	}

	// different XFF still allowed
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	req3.RemoteAddr = "127.0.0.1:9999"
	req3.Header.Set("X-Forwarded-For", "203.0.113.99")
	rec3 := httptest.NewRecorder()
	rl.Limit(okHandler()).ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("different XFF should be allowed, got %d", rec3.Code)
	}
}

func TestTokenBucket_UsesXRealIP(t *testing.T) {
	rl, mr := setupRateLimiter(t, 1, 0.001)
	defer mr.Close()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:9999"
	req.Header.Set("X-Real-IP", "198.51.100.7")
	rec := httptest.NewRecorder()
	rl.Limit(okHandler()).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("X-Real-IP request: %d", rec.Code)
	}

	// same X-Real-IP again → limited
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.RemoteAddr = "127.0.0.1:9999"
	req2.Header.Set("X-Real-IP", "198.51.100.7")
	rec2 := httptest.NewRecorder()
	rl.Limit(okHandler()).ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 for repeated X-Real-IP, got %d", rec2.Code)
	}
}

func TestTokenBucket_FailsOpenWhenRedisDown(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	rl := middleware.NewRateLimiterWithConfig(client, middleware.TokenBucketConfig{
		Capacity:   1,
		RefillRate: 0.001,
		KeyPrefix:  "tb:fail:",
		TTL:        time.Minute,
	})

	mr.Close()

	rec := doRequest(rl, "10.0.0.99")
	// current implementation fails open → 200
	if rec.Code != http.StatusOK {
		t.Fatalf("expected fail-open (200) when Redis is down, got %d", rec.Code)
	}
}
