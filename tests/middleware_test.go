package tests

import (
	"Linux-url-shortener/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddleware_CapturesStatus(t *testing.T) {
	handler := middleware.Logging(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/missing",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusNotFound,
			rec.Code,
		)
	}
}

func TestLoggingMiddleware_DefaultsTooK(t *testing.T) {
	handler := middleware.Logging(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("hello"))
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}

func TestMetricMiddleware(t *testing.T) {
	handler := middleware.MetricMiddleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/test-metrics",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected %d, got %d",
			http.StatusOK,
			rec.Code,
		)
	}
}
