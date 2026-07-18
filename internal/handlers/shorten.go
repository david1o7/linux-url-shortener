package handlers

import (
	"Linux-url-shortener/internal/cache"
	"Linux-url-shortener/internal/database"
	"Linux-url-shortener/internal/logger"
	"Linux-url-shortener/internal/services"
	"database/sql"
	"encoding/json"
	
	"net/http"
	"strings"
)

type Request struct{
	URL string `json:"url"`
}

type Response struct{
	ShortCode string `json:"short_code"`
}

func Shorten(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req Request

		err := json.NewDecoder(r.Body).Decode(&req)

		if err != nil{
			http.Error(w, "Invalid content", http.StatusBadRequest)
			return
		}
		code , err:= services.GenerateUniqueCode(db)
		if err != nil{
			http.Error(w, err.Error(), 500)
			return
		}
		err = database.SaveUrl(db, code, req.URL)
		if err != nil{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := Response{
			ShortCode: "http://localhost:8080/" + code,
		}
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(resp)

		logger.Log.Info(
			"Short URL Created",
			"Shortcode", code,
			"Original", req.URL,
		)
	}
}

func OriginalUrl(db *sql.DB, cache *cache.RedisCache) http.HandlerFunc {
	return func (w http.ResponseWriter, r *http.Request){
		code := strings.TrimPrefix(r.URL.Path, "/")

		url , err := cache.Get(code)

		if err == nil{
			logger.Log.Info(
				"cache hit",
				"Original Url", url,
			)
			
			logger.Log.Info(
				"Redirecting...",
				"Original Url", url,
			)

			go func(){
				err := database.IncrementClicks(db, code)

			if err != nil{
					logger.Log.Error(
						"Failed to increment clicks",
						"Shortcode", code,
						"error", err,
					)
				}
			}()

			http.Redirect(w,r,url, http.StatusFound)

			return
		}

		logger.Log.Info(
			"cache miss",
			"short code", url,
		)

		original, err := database.GetUrl(db, code)

		if err != nil || original == ""{
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cache.Set(code, original)

		
		logger.Log.Info(
				"Redirecting...",
				"Original Url", original,
			)
		
		go func(){
			err := database.IncrementClicks(db, code)

			if err != nil{
					logger.Log.Error(
						"Failed to increment clicks",
						"Shortcode", code,
						"error", err,
					)
				}
			}()

		http.Redirect(w, r, original, http.StatusFound)
	}
}