package tests

import (
	"Linux-url-shortener/internal/config"
	"os"
	"strings"
	"testing"
	"time"
)

func setEnv(t *testing.T, kv map[string]string) {
	t.Helper()
	for k, v := range kv {
		t.Setenv(k, v)
	}
}

func validEnv(t *testing.T) {
	t.Helper()
	setEnv(t, map[string]string{
		"DB_PASSWORD": "secret",
		"DB_HOST":     "localhost",
		"DB_NAME":     "url_shortener",
		"PORT":        "8080",
	})
}

func TestConfig_Load_Defaults(t *testing.T) {
	validEnv(t)
	// Clear optional overrides so defaults apply
	for _, k := range []string{
		"APP_ENV", "BASE_URL", "DB_PORT", "DB_USER", "DB_SSLMODE",
		"DB_MAX_OPEN_CONNS", "DB_MAX_IDLE_CONNS", "REDIS_ADDR",
		"RATE_LIMIT_CAPACITY", "RATE_LIMIT_REFILL_PER_SEC",
	} {
		_ = os.Unsetenv(k)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "8080" {
		t.Fatalf("Port = %q", cfg.Port)
	}
	if cfg.ShutDownTimeout != 5*time.Second {
		t.Fatalf("ShutDownTimeout = %v", cfg.ShutDownTimeout)
	}
	if cfg.DBMaxOpenConns != 25 {
		t.Fatalf("DBMaxOpenConns = %d", cfg.DBMaxOpenConns)
	}
	if cfg.RateLimitCapacity < 1 {
		t.Fatal("RateLimitCapacity should be >= 1")
	}
}

func TestConfig_Load_MissingPassword(t *testing.T) {
	_ = os.Unsetenv("DB_PASSWORD")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_NAME", "url_shortener")
	t.Setenv("PORT", "8080")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error for missing DB_PASSWORD")
	}
	if !strings.Contains(err.Error(), "DB_PASSWORD") {
		t.Fatalf("error = %v", err)
	}
}

func TestConfig_Validate_BadPool(t *testing.T) {
	cfg := &config.Config{
		DBPassword:            "x",
		DBHost:                "localhost",
		DBName:                "db",
		Port:                  "8080",
		DBMaxOpenConns:        0,
		DBMaxIdleConns:        0,
		RateLimitCapacity:     10,
		RateLimitRefillPerSec: 1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected DB_MAX_OPEN_CONNS error")
	}
}

func TestConfig_Validate_BadRateLimit(t *testing.T) {
	cfg := &config.Config{
		DBPassword:            "x",
		DBHost:                "localhost",
		DBName:                "db",
		Port:                  "8080",
		DBMaxOpenConns:        5,
		DBMaxIdleConns:        1,
		RateLimitCapacity:     0,
		RateLimitRefillPerSec: 1,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected RATE_LIMIT_CAPACITY error")
	}
}

func TestConfig_Validate_BadRefill(t *testing.T) {
	cfg := &config.Config{
		DBPassword:            "x",
		DBHost:                "localhost",
		DBName:                "db",
		Port:                  "8080",
		DBMaxOpenConns:        5,
		DBMaxIdleConns:        1,
		RateLimitCapacity:     10,
		RateLimitRefillPerSec: 0,
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected RATE_LIMIT_REFILL_PER_SEC error")
	}
}

func TestConfig_PostgresDSN(t *testing.T) {
	cfg := &config.Config{
		DBHost:     "h",
		DBPort:     "5432",
		DBUser:     "u",
		DBPassword: "p",
		DBName:     "n",
		DBSSLMode:  "disable",
	}
	dsn := cfg.PostgresDSN()
	for _, part := range []string{"host=h", "user=u", "password=p", "dbname=n"} {
		if !strings.Contains(dsn, part) {
			t.Fatalf("DSN missing %q: %s", part, dsn)
		}
	}
}

func TestConfig_IsProduction(t *testing.T) {
	cfg := &config.Config{Env: "production"}
	if !cfg.IsProduction() {
		t.Fatal("expected production")
	}
	cfg.Env = "Production"
	if !cfg.IsProduction() {
		t.Fatal("expected case-insensitive production")
	}
	cfg.Env = "development"
	if cfg.IsProduction() {
		t.Fatal("expected not production")
	}
}

func TestConfig_Load_InvalidNumbersFallBack(t *testing.T) {
	validEnv(t)
	t.Setenv("DB_MAX_OPEN_CONNS", "not-a-number")
	t.Setenv("RATE_LIMIT_REFILL_PER_SEC", "nope")
	t.Setenv("CONTEXT_TIMEOUT", "abc")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	// helpers fall back to defaults on parse error
	if cfg.DBMaxOpenConns != 25 {
		t.Fatalf("DBMaxOpenConns = %d, want default 25", cfg.DBMaxOpenConns)
	}
	if cfg.ShutDownTimeout != 5*time.Second {
		t.Fatalf("ShutDownTimeout = %v", cfg.ShutDownTimeout)
	}
}
