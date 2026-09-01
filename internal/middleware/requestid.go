package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type ctxKey int

const requestIDKey ctxKey = 1

const (
	HeaderRequestID     = "X-Request-ID"
	HeaderCorrelationID = "X-Correlation-ID"
)

func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderRequestID)
		if id == "" {
			id = r.Header.Get(HeaderCorrelationID)
		}
		if id == "" {
			id = newRequestID()
		}

		ctx := WithRequestID(r.Context(), id)
		w.Header().Set(HeaderRequestID, id)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Extremely unlikely; still return something non-empty.
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b)
}
