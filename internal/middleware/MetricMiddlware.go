package middleware

import (
	metrics "Linux-url-shortener/internal/metric"
	"net/http"
	"time"
)

func MetricMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		start := time.Now()

		next.ServeHTTP(w,r)

		metrics.RequestCounter.WithLabelValues(r.Method, r.URL.Path).Inc()

		metrics.RequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(time.Since(start).Seconds())

	})
}