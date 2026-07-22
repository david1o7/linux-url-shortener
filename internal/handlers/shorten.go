package handlers

import (
	"Linux-url-shortener/internal/cache"
	"Linux-url-shortener/internal/database"
	"Linux-url-shortener/internal/logger"
	metrics "Linux-url-shortener/internal/metric"
	"Linux-url-shortener/internal/validator"

	"Linux-url-shortener/internal/services"
	"database/sql"
	"encoding/json"
	"os"

	"net/http"
	"strings"

	"github.com/joho/godotenv"
)

type Request struct{
	URL string `json:"url"`
}

type Response struct{
	ShortCode string `json:"short_code"`
}

func Shorten(db *sql.DB, validator *validator.URLValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req Request

		err := godotenv.Load()

		if err != nil{
			logger.Log.Error(
				".env file error",
				"Error", err, 
			)
			http.Error(w, "Env file not found", http.StatusInternalServerError)
			
			return
		}

		BaseUrl := os.Getenv("BASE_URL")

		err = json.NewDecoder(r.Body).Decode(&req)

		if err != nil{
			http.Error(w, "Invalid content", http.StatusBadRequest)
			return
		}

		if !validator.Validate(req.URL){
			http.Error(w, "Invalid Url or cant be found", http.StatusBadRequest)
			metrics.InvalidUrls.Inc()
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

		metrics.UrlsShortened.Inc()

		resp := Response{
			ShortCode: BaseUrl + code,
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

			metrics.CacheHits.Inc()
			
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
			metrics.Redirects.Inc()
			return
		}

		logger.Log.Info(
			"cache miss",
			"short code", url,
		)
		metrics.CacheMisses.Inc()

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