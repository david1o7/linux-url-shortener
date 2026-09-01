package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"Linux-url-shortener/internal/handlers"
)

func TestIntegration_ShortenSuccess(t *testing.T) {
	resetState(t)
	h := handlers.Shorten(shared.Repo, alwaysPassValidator(), shared.CFG)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	var resp handlers.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(resp.ShortCode, shared.CFG.BaseURL+"/") {
		t.Fatalf("short_code=%q", resp.ShortCode)
	}

	code := resp.ShortCode[strings.LastIndex(resp.ShortCode, "/")+1:]
	got, err := shared.Repo.GetUrl(code)
	if err != nil {
		t.Fatal(err)
	}
	if got != "https://example.com" {
		t.Fatalf("db url=%q", got)
	}
}

func TestIntegration_ShortenInvalidURL(t *testing.T) {
	resetState(t)
	h := handlers.Shorten(shared.Repo, alwaysFailValidator(), shared.CFG)

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestIntegration_ShortenBadJSON(t *testing.T) {
	resetState(t)
	h := handlers.Shorten(shared.Repo, alwaysPassValidator(), shared.CFG)

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader("{"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestIntegration_RedirectCacheMissThenHit(t *testing.T) {
	resetState(t)

	t.Cleanup(func() {
		time.Sleep(100 * time.Millisecond) // let async IncrementClicks finish
	})

	if err := shared.Repo.SaveUrl("rdir01", "https://redirect.example"); err != nil {
		t.Fatal(err)
	}

	h := handlers.OriginalUrl(shared.Repo, shared.Cache)

	// 1) cache miss → DB → populate cache → 302
	req := httptest.NewRequest(http.MethodGet, "/rdir01", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("miss status=%d body=%s", rec.Code, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); loc != "https://redirect.example" {
		t.Fatalf("Location=%q", loc)
	}

	// cache should now have the key
	val, err := shared.Cache.Get("rdir01")
	if err != nil || val != "https://redirect.example" {
		t.Fatalf("cache after miss: %q err=%v", val, err)
	}

	// 2) cache hit → 302
	req2 := httptest.NewRequest(http.MethodGet, "/rdir01", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusFound {
		t.Fatalf("hit status=%d", rec2.Code)
	}

	// async IncrementClicks — give it a moment
	time.Sleep(150 * time.Millisecond)
	var clicks int
	_ = shared.DB.QueryRow(`SELECT clicks FROM urls WHERE shortcode = $1`, "rdir01").Scan(&clicks)
	if clicks < 1 {
		t.Fatalf("expected clicks >= 1 after redirects, got %d", clicks)
	}
}

func TestIntegration_RedirectMissingCode(t *testing.T) {
	resetState(t)

	t.Cleanup(func() {
		time.Sleep(100 * time.Millisecond) // let async IncrementClicks finish
	})
	h := handlers.OriginalUrl(shared.Repo, shared.Cache)

	req := httptest.NewRequest(http.MethodGet, "/nope99", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d (handler currently returns 500 on missing)", rec.Code)
	}
}

func TestIntegration_ShortenThenRedirect_E2E(t *testing.T) {
	resetState(t)

	t.Cleanup(func() {
		time.Sleep(100 * time.Millisecond) // let async IncrementClicks finish
	})

	shorten := handlers.Shorten(shared.Repo, alwaysPassValidator(), shared.CFG)
	redirect := handlers.OriginalUrl(shared.Repo, shared.Cache)

	body, _ := json.Marshal(map[string]string{"url": "https://e2e.example/path"})
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	shorten.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("shorten: %d %s", rec.Code, rec.Body.String())
	}

	var resp handlers.Response
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	code := resp.ShortCode[strings.LastIndex(resp.ShortCode, "/")+1:]

	req2 := httptest.NewRequest(http.MethodGet, "/"+code, nil)
	rec2 := httptest.NewRecorder()
	redirect.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusFound {
		t.Fatalf("redirect: %d", rec2.Code)
	}
	if loc := rec2.Header().Get("Location"); loc != "https://e2e.example/path" {
		t.Fatalf("Location=%q", loc)
	}
}
