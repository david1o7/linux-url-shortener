package main

import (
	// "fmt"
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"Linux-url-shortener/internal/cache"
	"Linux-url-shortener/internal/database"
	"Linux-url-shortener/internal/handlers"
	"Linux-url-shortener/internal/logger"
	metrics "Linux-url-shortener/internal/metric"
	"Linux-url-shortener/internal/middleware"
	"Linux-url-shortener/internal/validator"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main(){
	metrics.Init()

	err := godotenv.Load()

	if err != nil{
		logger.Log.Error(
			".Env file Error",
			"Error", err,
		)
	}

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	db_name := os.Getenv("DB_NAME")
	db_sslmode := os.Getenv("DB_SSLMODE")
	validatorTimeout, err := strconv.Atoi(os.Getenv("VALIDATOR_TIMEOUT"))
	if err != nil {
		validatorTimeout = 10
	}

	shutdownTimeout, err := strconv.Atoi(os.Getenv("CONTEXT_TIMEOUT"))
	if err != nil {
		shutdownTimeout = 5
	}
	ServerPort := os.Getenv("PORT")

	db , err := database.Connect(host, port , user, password, db_name, db_sslmode)
	if err != nil{
		panic(err)
	}

	logger.Log.Info(
		"DB connection status",
		"Connection" , "Successful",
	)
	resolver := validator.RealResolver{}

	urlValidator := validator.NewURLValidator(nil, &resolver, validatorTimeout)

	redisCache := cache.NewRedisCache()

	healthHandler := handlers.NewHealthHandler(db, redisCache.Client)

	Mux := http.NewServeMux()

	Mux.HandleFunc("/shorten", handlers.Shorten(db, urlValidator))
	Mux.HandleFunc("/", handlers.OriginalUrl(db, redisCache))
	Mux.HandleFunc("/health", healthHandler.Health)
	Mux.Handle("/metrics", promhttp.Handler())

	Limiter := middleware.NewRateLimiter(redisCache.Client)

	MetricHandler := middleware.MetricMiddleware(Mux)

	handler := Limiter.Limit(MetricHandler)

	server := &http.Server{
		Addr : ":" + ServerPort,
		Handler: handler,
	}

	go func(){

	logger.Log.Info(
		"Server started",
		"port", ":"+ServerPort,
	)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed{
		logger.Log.Error(
			"Server failed",
			"error", err.Error(),
		)
	}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(
		quit,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-quit

	logger.Log.Info("Shutdown signal received")

	ctx , cancel := context.WithTimeout(
		context.Background(),
		time.Duration(shutdownTimeout) * time.Second,
	)

	defer cancel()

	logger.Log.Info(
    "Waiting for active requests to finish...",
	)

	if err := server.Shutdown(ctx); err!= nil{

		logger.Log.Error(
			"Graceful shutdown failed",
			"error", err.Error(),
		)
	}

	if err := redisCache.Client.Close(); err != nil {

    logger.Log.Error(
        "Redis close failed",
        "error",
        err.Error(),
    )
	} else {
    logger.Log.Info("Redis connection closed")
	}

	if err := db.Close(); err != nil {

    logger.Log.Error(
        "Database close failed",
        "error",
        err.Error(),
    )

	} else {
    logger.Log.Info("DB connection closed")
	}

	logger.Log.Info("Server shutdown complete")
}