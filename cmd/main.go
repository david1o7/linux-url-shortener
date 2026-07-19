package main

import (
	// "fmt"
	"net/http"
	"os"

	"Linux-url-shortener/internal/cache"
	"Linux-url-shortener/internal/database"
	"Linux-url-shortener/internal/handlers"
	"Linux-url-shortener/internal/logger"
	"Linux-url-shortener/internal/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/joho/godotenv"
)

func main(){
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

	ServerPort := os.Getenv("PORT")

	db , err := database.Connect(host, port , user, password, db_name, db_sslmode)
	if err != nil{
		panic(err)
	}

	logger.Log.Info(
		"DB connection status",
		"Connection" , "Successful",
	)

	cache := cache.NewRedisCache()

	healthHandler := handlers.NewHealthHandler(db, cache.Client)

	Mux := http.NewServeMux()

	Mux.HandleFunc("/shorten", handlers.Shorten(db))
	Mux.HandleFunc("/", handlers.OriginalUrl(db, cache))
	Mux.HandleFunc("/health", healthHandler.Health)
	Mux.Handle("/metrics", promhttp.Handler())

	Limiter := middleware.NewRateLimiter(cache.Client)

	server := Limiter.Limit(Mux)

	logger.Log.Info(
		"Server started",
		"port", ":"+ServerPort,
	)
	http.ListenAndServe(":"+ServerPort, server)
}