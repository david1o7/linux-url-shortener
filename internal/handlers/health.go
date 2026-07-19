package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type HealthHandler struct{
	DB *sql.DB
	Redis *redis.Client
}

type HealthStatus struct{
	Status string `json:"status"`
	Database string `json:"db"`
	Redis string `json:"redis"`
}

func NewHealthHandler(db *sql.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		DB: db,
		Redis: redis,
	}
}

func (h *HealthHandler) Health(w http.ResponseWriter , r *http.Request) {

	response := HealthStatus{
		Status: "healthy",
		Database: "up",
		Redis: "up",
	}

	if err := h.DB.Ping(); err != nil{
		response.Database = "down"
		response.Status = "unhealthy"
	}

	if err := h.Redis.Ping(r.Context()).Err(); err != nil{
		response.Status = "unhealthy"
		response.Redis = "down"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}