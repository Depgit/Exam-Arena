package database

import (
	"context"
	"database/sql"
	"strings"

	"github.com/exam-arena/internal/config"
)

// InitDB initializes a database connection (*sql.DB) based on configuration (SQLite or PostgreSQL).
func InitDB(ctx context.Context, cfg *config.Config) (*sql.DB, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.DatabaseDriver))
	switch driver {
	case "postgres", "postgresql", "pgx":
		return NewPostgresSQLDB(ctx, cfg.DatabaseURL, cfg.DBMaxConns, cfg.DBMinConns)
	default:
		return NewSQLite(ctx, cfg.DatabaseURL)
	}
}
