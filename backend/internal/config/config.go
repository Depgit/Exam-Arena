package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerHost          string
	ServerPort          string
	Environment         string
	DatabaseURL         string
	DBMaxConns          int32
	DBMinConns          int32
	JWTSecret           string
	JWTExpiryHours      int
	RateLimitRPS        float64
	RateLimitBurst      int
	CORSAllowedOrigins  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		ServerHost:         getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		DatabaseURL:        getEnv("DATABASE_URL", "exam_arena.db"),
		JWTSecret:          getEnv("JWT_SECRET", "drpzet_default_secret_key_123"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
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
