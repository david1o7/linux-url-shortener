package middleware

import (
	"Linux-url-shortener/internal/logger"
	metrics "Linux-url-shortener/internal/metric"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenBucketConfig struct {
	Capacity   int64
	RefillRate float64
	KeyPrefix  string
	TTL        time.Duration
}

func DefualtConfig() TokenBucketConfig {
	return TokenBucketConfig{
		Capacity:   20,
		RefillRate: 1.0,
		KeyPrefix:  "tb:",
		TTL:        2 * time.Minute,
	}
}

type RateLimiter struct {
	Client *redis.Client
	Cfg    TokenBucketConfig
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{
		Client: client,
		Cfg:    DefualtConfig(),
	}
}

// Optional: for custom config, tests
func NewRateLimiterWithConfig(client *redis.Client, cfg TokenBucketConfig) *RateLimiter {
	return &RateLimiter{
		Client: client,
		Cfg:    cfg,
	}
}

var tokenBucketScript = redis.NewScript(
	`
local key          = KEYS[1]
local capacity     = tonumber(ARGV[1])
local refill_rate  = tonumber(ARGV[2])
local now          = tonumber(ARGV[3])
local requested    = tonumber(ARGV[4])
local ttl          = tonumber(ARGV[5])

local data = redis.call("HMGET", key, "tokens", "last_refill")
local tokens = tonumber(data[1])
local last_refill = tonumber(data[2])

if tokens == nil then
    tokens = capacity
    last_refill = now
end

-- Refill based on elapsed time
local elapsed = now - last_refill
if elapsed > 0 then
    local added = elapsed * refill_rate
    tokens = math.min(capacity, tokens + added)
    last_refill = now
end

local allowed = 0
local retry_after = 0

if tokens >= requested then
    tokens = tokens - requested
    allowed = 1
else
    -- How long until we have enough tokens?
    local deficit = requested - tokens
    retry_after = math.ceil(deficit / refill_rate)
end

redis.call("HMSET", key, "tokens", tokens, "last_refill", last_refill)
redis.call("EXPIRE", key, ttl)

return {allowed, tokens, retry_after}
`)

func (r *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()

		ip := extractIP(req)
		key := r.Cfg.KeyPrefix + ip

		now := float64(time.Now().UnixNano()) / 1e9

		result, err := tokenBucketScript.Run(ctx, r.Client,
			[]string{key},
			r.Cfg.Capacity,
			r.Cfg.RefillRate,
			now,
			1,
			int(r.Cfg.TTL.Seconds()),
		).Result()

		if err != nil {
			logger.Log.Error(
				"Token Bucket Redis error",
				"Error", err,
				"ip", ip,
			)

			next.ServeHTTP(w, req)
			return
		}

		vals, ok := result.([]interface{})
		if !ok || len(vals) < 3 {
			logger.Log.Error(
				"Unexpected token bucket script result",
				"result", result,
			)

			next.ServeHTTP(w, req)
			return
		}

		allowed, _ := vals[0].(int64)
		remaining, _ := vals[1].(int64)
		retryAfter, _ := vals[2].(int64)

		w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(r.Cfg.Capacity, 10))
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(int64(remaining), 10))
		if allowed == 0 {
			w.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))
		}

		if allowed == 0 {
			logger.Log.Warn(
				"Rate limit exceeded (token bucket)",
				"ip", ip,
				"remaining", remaining,
			)
			metrics.RateLimited.Inc()
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, req)

	})
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {

		if i := len(xff); i > 0 {
			for j := 0; j < i; j++ {
				if xff[j] == ',' {
					return xff[:j]
				}
			}
			return xff
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
