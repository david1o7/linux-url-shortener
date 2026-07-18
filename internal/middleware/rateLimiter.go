package middleware

import (
	"Linux-url-shortener/internal/logger"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct{
	Client *redis.Client
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	return &RateLimiter{
		Client: client,
	}
}

func (r *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request){
		ctx := req.Context()

		ip := req.RemoteAddr

		key := "rate:" + ip

		count, err := r.Client.Incr(ctx, key).Result()

		if err != nil{
			http.Error(w, "Redis Error", 500)
			return
		}

		if count == 1{
			r.Client.Expire(ctx, key, time.Minute)
		}

		if count > 100 {
			
		logger.Log.Warn(
			"Rate Limit Exceeded",
			"ip", req.RemoteAddr,
		)
			http.Error(w, "Too Many Request", http.StatusTooManyRequests)
				return
		}
		next.ServeHTTP(w,req)
	})
}