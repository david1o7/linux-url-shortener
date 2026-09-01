package integration

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"Linux-url-shortener/internal/cache"
	"Linux-url-shortener/internal/config"
	"Linux-url-shortener/internal/database"
	"Linux-url-shortener/internal/validator"
	"Linux-url-shortener/tests/mocks"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// shared holds process-wide dependencies for the integration package.
// It is initialized once in TestMain and read by tests.
var shared struct {
	CFG   *config.Config
	DB    *sql.DB
	Cache *cache.RedisCache
	Repo  database.Repository

	pgContainer    testcontainers.Container
	redisContainer testcontainers.Container
}

// TestMain runs once for this package.
// Flow: start containers → connect → migrate → run tests → teardown.
func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	code := 1
	defer func() {
		// Always attempt cleanup, even if setup failed mid-way.
		shutdown(context.Background())
		os.Exit(code)
	}()

	if err := startContainers(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "integration setup failed: %v\n", err)
		return
	}
	if err := connectAndMigrate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "integration connect/migrate failed: %v\n", err)
		return
	}

	code = m.Run()
}

func startContainers(ctx context.Context) error {
	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("url_shortener"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return fmt.Errorf("postgres: %w", err)
	}
	shared.pgContainer = pg

	rd, err := redis.Run(ctx, "redis:7-alpine")
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	shared.redisContainer = rd
	return nil
}

func connectAndMigrate(ctx context.Context) error {
	pgHost, err := shared.pgContainer.Host(ctx)
	if err != nil {
		return err
	}
	pgPort, err := shared.pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		return err
	}
	redisEndpoint, err := shared.redisContainer.Endpoint(ctx, "")
	if err != nil {
		return err
	}

	cfg := &config.Config{
		Env:                   "test",
		Port:                  "8080",
		BaseURL:               "http://localhost:8080",
		ShutDownTimeout:       5 * time.Second,
		DBHost:                pgHost,
		DBPort:                pgPort.Port(),
		DBUser:                "postgres",
		DBPassword:            "test",
		DBName:                "url_shortener",
		DBSSLMode:             "disable",
		DBMaxOpenConns:        10,
		DBMaxIdleConns:        2,
		DBConnMaxLifetime:     2 * time.Minute,
		DBConnMaxIdleTime:     time.Minute,
		RedisAddr:             redisEndpoint,
		ValidatorTimeout:      2 * time.Second,
		RateLimitCapacity:     100,
		RateLimitRefillPerSec: 10,
		RateLimitTTL:          2 * time.Minute,
	}
	shared.CFG = cfg

	db, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	shared.DB = db

	if err := database.WaitForDB(db, 30*time.Second); err != nil {
		return fmt.Errorf("wait for db: %w", err)
	}
	if err := database.Migrate(db, migrationsDir()); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	rc := cache.NewRedisCache(cfg)
	shared.Cache = rc
	shared.Repo = database.NewPostgresRepository(db)
	return nil
}

func shutdown(ctx context.Context) {
	if shared.Cache != nil {
		_ = shared.Cache.Close()
	}
	if shared.DB != nil {
		_ = shared.DB.Close()
	}
	if shared.redisContainer != nil {
		_ = shared.redisContainer.Terminate(ctx)
	}
	if shared.pgContainer != nil {
		_ = shared.pgContainer.Terminate(ctx)
	}
}

func migrationsDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "migrations"
	}
	// tests/integration → <repo>/migrations
	return filepath.Join(filepath.Dir(file), "..", "..", "migrations")
}

func resetState(t *testing.T) {
	t.Helper()
	if shared.DB == nil {
		t.Fatal("shared DB not initialized — TestMain setup failed?")
	}

	// Order: truncate tables that reference others first if you add FKs later.
	if _, err := shared.DB.Exec(`TRUNCATE TABLE urls RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("truncate urls: %v", err)
	}

	// Flush Redis so cache hits from a previous test cannot affect this one.
	if shared.Cache != nil && shared.Cache.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := shared.Cache.Client.FlushDB(ctx).Err(); err != nil {
			t.Fatalf("redis flush: %v", err)
		}
	}
}

// --- validators (offline-safe) ---

func alwaysPassValidator() *validator.URLValidator {
	client := &mocks.MockClient{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(http.NoBody),
		},
	}
	resolver := &mocks.MockResolver{
		IPs: []net.IP{net.ParseIP("93.184.216.34")},
	}
	return validator.NewURLValidator(client, resolver, 2)
}

func alwaysFailValidator() *validator.URLValidator {
	client := &mocks.MockClient{Err: io.ErrUnexpectedEOF}
	resolver := &mocks.MockResolver{
		IPs: []net.IP{net.ParseIP("93.184.216.34")},
	}
	return validator.NewURLValidator(client, resolver, 2)
}
