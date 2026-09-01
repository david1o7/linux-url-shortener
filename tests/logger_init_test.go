package tests

import (
	"Linux-url-shortener/internal/logger"
	"testing"
)

func TestLogger_InitProductionAndDev(t *testing.T) {
	logger.Init("production")
	logger.Log.Info("prod-style")
	logger.Init("development")
	logger.Log.Info("dev-style")
}
