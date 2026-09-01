package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"Linux-url-shortener/internal/cache"
	"Linux-url-shortener/internal/config"
	"Linux-url-shortener/internal/database"
	"Linux-url-shortener/internal/handlers"
	"Linux-url-shortener/internal/logger"
	metrics "Linux-url-shortener/internal/metric"
	"Linux-url-shortener/internal/middleware"
	"Linux-url-shortener/internal/validator"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 1. Load & validate configuration
	cfg, err := config.Load()
	if err != nil {
		if n, err := os.Stderr.WriteString("config error: " + err.Error() + "\n"); err != nil {
			logger.Log.Error(
				"Error while listening for config errors",
				"Issues", n,
				"Error", err,
			)
		}
		os.Exit(1)
	}

	// 2. Structured logging (JSON in production)
	logger.Init(cfg.Env)
	logger.Log.Info("Starting URL shortener", "env", cfg.Env, "port", cfg.Port)

	metrics.Init()

	// 3. Database
	db, err := database.Connect(cfg)
	if err != nil {
		logger.Log.Error("Database connection failed", "error", err)
		os.Exit(1)
	}

	// Wait briefly for Postgres in container environments
	if err := database.WaitForDB(db, 30*time.Second); err != nil {
		logger.Log.Error("Database not ready", "error", err)
		os.Exit(1)
	}
	logger.Log.Info("Database connection successful")

	// 4. Run migrations
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := database.Migrate(db, migrationsDir); err != nil {
		logger.Log.Error("Migration failed", "error", err)
		os.Exit(1)
	}
	logger.Log.Info("Migrations up to date")

	// 5. Redis
	redisCache := cache.NewRedisCache(cfg)
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := redisCache.Ping(pingCtx); err != nil {
		pingCancel()
		logger.Log.Error("Redis connection failed", "error", err)
		os.Exit(1)
	}
	pingCancel()
	logger.Log.Info("Redis connection successful", "addr", cfg.RedisAddr)

	// 6. Dependencies
	resolver := validator.RealResolver{}
	urlValidator := validator.NewURLValidator(nil, &resolver, int(cfg.ValidatorTimeout.Seconds()))
	healthHandler := handlers.NewHealthHandler(db, redisCache.Client)
	repo := database.NewPostgresRepository(db)

	// 7. Routes
	mux := http.NewServeMux()
	mux.HandleFunc("/shorten", handlers.Shorten(repo, urlValidator, cfg))
	mux.HandleFunc("/", handlers.OriginalUrl(repo, redisCache))
	mux.HandleFunc("/health", healthHandler.Health)
	mux.HandleFunc("/health/live", healthHandler.Liveness)
	mux.HandleFunc("/health/ready", healthHandler.Readiness)
	mux.Handle("/metrics", promhttp.Handler())

	// 8. Middleware stack: request ID → logging → rate limit → metrics → routes
	TkB := middleware.TokenBucketConfig{
		Capacity:   int64(cfg.RateLimitCapacity),
		RefillRate: cfg.RateLimitRefillPerSec,
		KeyPrefix:  "tb:",
		TTL:        cfg.RateLimitTTL,
	}
	limiter := middleware.NewRateLimiterWithConfig(redisCache.Client, TkB)
	metricHandler := middleware.MetricMiddleware(mux)
	rateLimited := limiter.Limit(metricHandler)
	logged := middleware.Logging(rateLimited)
	handler := middleware.RequestID(logged)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 9. Start server
	go func() {
		logger.Log.Info("Server started", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// 10. Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		cfg.ShutDownTimeout,
	)
	defer shutdownCancel()

	logger.Log.Info("Waiting for active requests to finish...")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("Graceful shutdown failed", "error", err)
	}

	_ = redisCache.Close()

	if err := db.Close(); err != nil {
		logger.Log.Error("Database close failed", "error", err)
	} else {
		logger.Log.Info("Database connection closed")
	}

	logger.Log.Info("Server shutdown complete")
}
