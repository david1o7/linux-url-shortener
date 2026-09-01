package handlers

import (
	"Linux-url-shortener/internal/cache"
	"Linux-url-shortener/internal/config"
	"Linux-url-shortener/internal/database"
	"Linux-url-shortener/internal/logger"
	metrics "Linux-url-shortener/internal/metric"
	"Linux-url-shortener/internal/validator"

	"Linux-url-shortener/internal/services"
	"encoding/json"

	"net/http"
	"strings"
)

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	ShortCode string `json:"short_code"`
}

func Shorten(repo database.Repository, validator *validator.URLValidator, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var req Request

		BaseUrl := cfg.BaseURL

		err := json.NewDecoder(r.Body).Decode(&req)

		if err != nil {
			http.Error(w, "Invalid content", http.StatusBadRequest)
			return
		}

		if !validator.Validate(req.URL) {
			http.Error(w, "Invalid Url or cant be found", http.StatusBadRequest)
			metrics.InvalidUrls.Inc()
			return
		}

		code, err := services.GenerateUniqueCode(repo, req.URL, 10)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		metrics.UrlsShortened.Inc()

		resp := Response{
			ShortCode: BaseUrl + "/" + code,
		}
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		logger.Log.Info(
			"Short URL Created",
			"Shortcode", code,
			"Original", req.URL,
		)
	}
}

func OriginalUrl(repo database.Repository, cache *cache.RedisCache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := strings.TrimPrefix(r.URL.Path, "/")

		url, err := cache.Get(code)

		if err == nil {
			logger.Log.Info(
				"cache hit",
				"Original Url", url,
			)

			metrics.CacheHits.Inc()

			logger.Log.Info(
				"Redirecting...",
				"Original Url", url,
			)

			go func() {
				err := repo.IncrementClicks(code)

				if err != nil {
					logger.Log.Error(
						"Failed to increment clicks",
						"Shortcode", code,
						"error", err,
					)
				}
			}()

			http.Redirect(w, r, url, http.StatusFound)
			metrics.Redirects.Inc()
			return
		}

		logger.Log.Info(
			"cache miss",
			"short code", url,
		)
		metrics.CacheMisses.Inc()

		original, err := repo.GetUrl(code)

		if err != nil || original == "" {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = cache.Set(code, original); err != nil {
			logger.Log.Error(
				"Cache update Error",
				"Err", err,
			)
			return
		}

		logger.Log.Info(
			"Redirecting...",
			"Original Url", original,
		)

		go func() {
			err := repo.IncrementClicks(code)

			if err != nil {
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
