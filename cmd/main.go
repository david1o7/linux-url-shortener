package main

import (
	"fmt"
	"net/http"

	"Linux-url-shortener/internal/cache"
	"Linux-url-shortener/internal/database"
	"Linux-url-shortener/internal/handlers"
	"Linux-url-shortener/internal/logger"
	"Linux-url-shortener/internal/middleware"
)

func main(){
	db , err := database.Connect("172.17.160.1", "5432" , "postgres" , "Admin", "NewUrl_Shortener", "disable")
	if err != nil{
		panic(err)
	}
	fmt.Println("Connected to DB successfully!!")

	cache := cache.NewRedisCache()

	Mux := http.NewServeMux()

	Mux.HandleFunc("/shorten", handlers.Shorten(db))
	Mux.HandleFunc("/", handlers.OriginalUrl(db, cache))

	Limiter := middleware.NewRateLimiter(cache.Client)

	server := Limiter.Limit(Mux)

	logger.Log.Info(
		"Server started",
		"post", ":8080",
	)
	http.ListenAndServe(":8080", server)
}