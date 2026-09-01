package middleware

import (
	"Linux-url-shortener/internal/logger"
	"net/http"
	"time"
)

type responsewriter struct {
	http.ResponseWriter
	status int
}

func (rw *responsewriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responsewriter{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		attrs := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration", time.Since(start),
			"remote_addr", r.RemoteAddr,
		}
		if id := RequestIDFromContext(r.Context()); id != "" {
			attrs = append(attrs, "request_id", id)
		}

		logger.Log.Info("HTTP Request", attrs...)

	})
}
