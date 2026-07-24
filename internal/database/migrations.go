package database

import (
	"context"
	_ "embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

// / go:embed ../../migrations/001_initial_schema.sql
var initialSchema string

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Create migrations tracking table
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		return err
	}

	// Check if migration 1 was already applied
	var exists bool
	err = pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = 1)`).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil // Already migrated
	}

	// Run the migration
	_, err = pool.Exec(ctx, initialSchema)
	if err != nil {
		return err
	}

	// Record the migration
	_, err = pool.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES (1)`)
	return err
}
