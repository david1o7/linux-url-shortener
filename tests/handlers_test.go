package tests

import (
	"Linux-url-shortener/internal/cache"
	"Linux-url-shortener/internal/config"

	"Linux-url-shortener/internal/handlers"
	"Linux-url-shortener/internal/validator"
	"Linux-url-shortener/tests/mocks"
	"bytes"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func passingValidator() *validator.URLValidator {

	client := &mocks.MockClient{
		Response: &http.Response{StatusCode: 200, Body: http.NoBody},
	}

	return validator.NewURLValidator(client, &mocks.MockResolver{IPs: []net.IP{
		net.ParseIP("93.184.216.34"),
	}}, 2)
}

func failingValidator() *validator.URLValidator {
	client := &mocks.MockClient{Err: errors.New("boom")}

	return validator.NewURLValidator(client, &mocks.MockResolver{IPs: []net.IP{
		net.ParseIP("93.184.216.34"),
	}}, 2)
}

func testConfig() *config.Config {
	return &config.Config{
		BaseURL: "http://localhost:8080",
	}
}

func TestShorten_InvalidJSON(t *testing.T) {
	repo := &mocks.MockRepository{}
	h := handlers.Shorten(repo, passingValidator(), testConfig())

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader("{bad"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestShorten_InvalidURL(t *testing.T) {
	repo := &mocks.MockRepository{}
	h := handlers.Shorten(repo, failingValidator(), testConfig())

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestShorten_Success(t *testing.T) {
	repo := &mocks.MockRepository{}
	h := handlers.Shorten(repo, passingValidator(), testConfig())

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !repo.SaveURLCalled {
		t.Fatal("expected SaveUrl to be called via GenerateUniqueCode")
	}
	var resp handlers.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(resp.ShortCode, "http://localhost:8080/") {
		t.Fatalf("short_code = %q", resp.ShortCode)
	}
}

func TestShorten_RepoError(t *testing.T) {
	repo := &mocks.MockRepository{SaveErr: mocks.ErrDatabase}
	h := handlers.Shorten(repo, passingValidator(), testConfig())

	body, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	req := httptest.NewRequest(http.MethodPost, "/shorten", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestOriginalUrl_CacheHit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &cache.RedisCache{Client: client}
	_ = rc.Set("abc123", "https://example.com")

	repo := &mocks.MockRepository{}
	h := handlers.OriginalUrl(repo, rc)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "https://example.com" {
		t.Fatalf("Location = %q", loc)
	}
}

func TestOriginalUrl_CacheMiss_DBHit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &cache.RedisCache{Client: client}

	repo := &mocks.MockRepository{URL: "https://from-db.com"}
	h := handlers.OriginalUrl(repo, rc)

	req := httptest.NewRequest(http.MethodGet, "/xyz789", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !repo.GetCalled {
		t.Fatal("expected DB GetUrl")
	}
	if loc := rec.Header().Get("Location"); loc != "https://from-db.com" {
		t.Fatalf("Location = %q", loc)
	}
}

func TestOriginalUrl_NotFound(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rc := &cache.RedisCache{Client: client}

	// Prefer a plain error:
	repo := &mocks.MockRepository{GetErr: errors.New("sql: no rows")}
	h := handlers.OriginalUrl(repo, rc)

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestHealth_Liveness(t *testing.T) {
	h := handlers.NewHealthHandler(nil, nil)
	// Liveness does not touch DB/Redis
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()
	h.Liveness(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "alive") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}
