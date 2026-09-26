package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerHost     string
	ServerPort     string
	Environment    string
	DatabaseURL    string
	DBMaxConns     int32
	DBMinConns     int32
	JWTSecret      string
	JWTExpiryHours int
	RateLimitRPS   float64
	RateLimitBurst int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		ServerHost:  getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	cfg.DBMaxConns = int32(getEnvInt("DB_MAX_CONNS", 25))
	cfg.DBMinConns = int32(getEnvInt("DB_MIN_CONNS", 5))
	cfg.JWTExpiryHours = getEnvInt("JWT_EXPIRY_HOURS", 24)
	cfg.RateLimitRPS = getEnvFloat("RATE_LIMIT_RPS", 100)
	cfg.RateLimitBurst = getEnvInt("RATE_LIMIT_BURST", 200)

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}
