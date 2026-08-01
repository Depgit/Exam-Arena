package database

import (
	"context"

	"database/sql"
	_ "embed"
)

//go:embed schema.sql
var initialSchema string

func RunMigrations(ctx context.Context, db *sql.DB) error {
	// Create migrations tracking table
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Check if migration 1 was already applied
	var versionCount int
	err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version = 1`).Scan(&versionCount)
	if err != nil {
		return err
	}

	if versionCount > 0 {
		return nil // Already migrated
	}

	// Run the migration schema
	_, err = db.ExecContext(ctx, initialSchema)
	if err != nil {
		return err
	}

	// Record the migration
	_, err = db.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (1)`)
	return err
}
