package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

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

	// Question generator (internal/questiongen, questionpool.Pool).
	QuestionGenEnabled        bool
	QuestionGenPoolSize       int
	QuestionGenRotateEvery    time.Duration
	QuestionGenPerUser        int
	QuestionGenMaxPerCategory int // ceiling on questions per category (the bank lives in memory)

	// Sign-in (internal/auth).
	GoogleClientID string // "Sign in with Google" web client id; "" = off
	ResendAPIKey   string // email via resend.com; "" = emails are only logged
	MailFrom       string
	// EmailVerification: "required" (verify before playing) or "off".
	// Default: required when ResendAPIKey is set, otherwise off.
	EmailVerification string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		ServerHost: getEnv("SERVER_HOST", "0.0.0.0"),
		// Render (and most PaaS hosts) inject PORT and route traffic to it;
		// SERVER_PORT still wins when set explicitly.
		ServerPort:  getEnv("SERVER_PORT", getEnv("PORT", "8080")),
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

	cfg.QuestionGenEnabled = getEnvBool("QUESTION_GEN_ENABLED", false)
	cfg.QuestionGenPoolSize = getEnvInt("QUESTION_GEN_POOL_SIZE", 600)
	cfg.QuestionGenRotateEvery = time.Duration(getEnvInt("QUESTION_GEN_ROTATE_HOURS", 6)) * time.Hour
	cfg.QuestionGenPerUser = getEnvInt("QUESTION_GEN_PER_USER", 100)
	cfg.QuestionGenMaxPerCategory = getEnvInt("QUESTION_GEN_MAX_PER_CATEGORY", 5000)

	cfg.GoogleClientID = getEnv("GOOGLE_CLIENT_ID", "")
	cfg.ResendAPIKey = getEnv("RESEND_API_KEY", "")
	cfg.MailFrom = getEnv("MAIL_FROM", "Mind Race <noreply@mindrace.in>")
	defaultVerification := "off"
	if cfg.ResendAPIKey != "" {
		defaultVerification = "required"
	}
	cfg.EmailVerification = getEnv("EMAIL_VERIFICATION", defaultVerification)

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

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
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
