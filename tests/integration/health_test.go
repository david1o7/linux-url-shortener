package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Linux-url-shortener/internal/handlers"
)

func TestIntegration_HealthLiveReady(t *testing.T) {
	resetState(t)
	h := handlers.NewHealthHandler(shared.DB, shared.Cache.Client)

	// Liveness — no dependency checks
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	h.Liveness(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("live status=%d", rec.Code)
	}

	// Readiness — DB + Redis up
	req2 := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec2 := httptest.NewRecorder()
	h.Readiness(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("ready status=%d body=%s", rec2.Code, rec2.Body.String())
	}

	var body handlers.HealthStatus
	if err := json.Unmarshal(rec2.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ready" || body.Database != "up" || body.Redis != "up" {
		t.Fatalf("unexpected readiness body: %+v", body)
	}

	// Combined health
	req3 := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec3 := httptest.NewRecorder()
	h.Health(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("health status=%d", rec3.Code)
	}
}
