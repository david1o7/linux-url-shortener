package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env string

	Port            string
	ShutDownTimeout time.Duration
	BaseURL         string

	DBHost            string
	DBPort            string
	DBUser            string
	DBName            string
	DBPassword        string
	DBSSLMode         string
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	ValidatorTimeout time.Duration

	RateLimitCapacity     int
	RateLimitRefillPerSec float64
	RateLimitTTL          time.Duration
}

func getdurationseconds(str string, def int) time.Duration {
	val := os.Getenv(str)
	if val == "" {
		return time.Duration(def) * time.Second
	}
	t, err := strconv.Atoi(val)
	if err != nil {
		t = def
	}
	return time.Duration(t) * time.Second
}

func getInt(str string, def int) int {
	val := os.Getenv(str)
	if val == "" {
		return def
	}
	t, err := strconv.Atoi(val)
	if err != nil {
		t = def
	}
	return t
}

func getFloat(str string, def float64) float64 {
	val := os.Getenv(str)
	if val == "" {
		return def
	}
	t, err := strconv.ParseFloat(val, 64)
	if err != nil {
		t = def
	}
	return t
}

func getEnv(str string, def string) string {
	if v := os.Getenv(str); v != "" {
		return v
	}
	return def
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Env: getEnv("APP_ENV", "development"),

		Port:            getEnv("PORT", "8080"),
		ShutDownTimeout: getdurationseconds("CONTEXT_TIMEOUT", 5),
		BaseURL:         getEnv("BASE_URL", "http://localhost:8080"),

		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        os.Getenv("DB_PASSWORD"),
		DBName:            getEnv("DB_NAME", "NewUrl_Shortener"),
		DBSSLMode:         getEnv("DB_SSLMODE", "disable"),
		DBMaxOpenConns:    getInt("DB_MAX_OPEN_CONNS", 25),
		DBMaxIdleConns:    getInt("DB_MAX_IDLE_CONNS", 5),
		DBConnMaxLifetime: getdurationseconds("DB_CONN_MAX_LIFETIME_SEC", 300),
		DBConnMaxIdleTime: getdurationseconds("DB_CONN_MAX_IDLE_TIME_SEC", 60),

		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       getInt("REDIS_DB", 0),

		ValidatorTimeout: getdurationseconds("VALIDATOR_TIMEOUT", 10),

		RateLimitCapacity:     getInt("RATE_LIMIT_CAPACITY", 20),
		RateLimitRefillPerSec: getFloat("RATE_LIMIT_REFILL_PER_SEC", 1.0),
		RateLimitTTL:          getdurationseconds("RATE_LIMIT_TTL_SEC", 120),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	var missing []string

	if c.DBPassword == "" {
		missing = append(missing, "DB_PASSWORD")
	}
	if c.DBHost == "" {
		missing = append(missing, "DB_HOST")
	}
	if c.DBName == "" {
		missing = append(missing, "DB_NAME")
	}
	if c.Port == "" {
		missing = append(missing, "PORT")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	if c.DBMaxOpenConns < 1 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS must be >= 1")
	}
	if c.DBMaxIdleConns < 0 {
		return fmt.Errorf("DB_MAX_IDLE_CONNS must be >= 0")
	}
	if c.RateLimitCapacity < 1 {
		return fmt.Errorf("RATE_LIMIT_CAPACITY must be >= 1")
	}
	if c.RateLimitRefillPerSec <= 0 {
		return fmt.Errorf("RATE_LIMIT_REFILL_PER_SEC must be > 0")
	}

	return nil
}

func (c *Config) PostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.Env, "production")
}
