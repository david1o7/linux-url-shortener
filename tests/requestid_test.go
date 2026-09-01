package tests

import (
	"Linux-url-shortener/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	var got string
	h := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = middleware.RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got == "" {
		t.Fatal("expected request id in context")
	}
	if rec.Header().Get(middleware.HeaderRequestID) == "" {
		t.Fatal("expected X-Request-ID response header")
	}
	if rec.Header().Get(middleware.HeaderRequestID) != got {
		t.Fatalf("header %q != context %q", rec.Header().Get(middleware.HeaderRequestID), got)
	}
}

func TestRequestID_PropagatesIncomingHeader(t *testing.T) {
	const want = "client-provided-id-123"

	var got string
	h := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = middleware.RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/shorten", nil)
	req.Header.Set(middleware.HeaderRequestID, want)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got != want {
		t.Fatalf("context id = %q, want %q", got, want)
	}
	if rec.Header().Get(middleware.HeaderRequestID) != want {
		t.Fatalf("response header = %q, want %q", rec.Header().Get(middleware.HeaderRequestID), want)
	}
}

func TestRequestID_AcceptsCorrelationHeader(t *testing.T) {
	const want = "corr-xyz"

	var got string
	h := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = middleware.RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(middleware.HeaderCorrelationID, want)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got != want {
		t.Fatalf("context id = %q, want %q", got, want)
	}
}

func TestRequestID_LoggingStackStillOK(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if middleware.RequestIDFromContext(r.Context()) == "" {
			t.Error("missing request id inside handler")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	h := middleware.RequestID(middleware.Logging(inner))

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Header().Get(middleware.HeaderRequestID) == "" {
		t.Fatal("expected X-Request-ID on response")
	}
}
