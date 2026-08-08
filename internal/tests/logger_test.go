package tests

import (
	"Linux-url-shortener/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
	
)

func TestLoggingMiddleware(t *testing.T) {
	handler := middleware.Logging(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet, "/health", nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {

		t.Fatalf("expected %d gor %d", http.StatusOK, rec.Code)
	}
}
