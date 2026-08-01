package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// NewSQLite initializes a connection to a local SQLite database file.
func NewSQLite(ctx context.Context, dbPath string) (*sql.DB, error) {
	if dbPath == "" {
		dbPath = "exam_arena.db"
	}

	// Enable WAL mode and foreign keys for high concurrent performance
	connStr := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)", dbPath)

	db, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, fmt.Errorf("open sqlite db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping sqlite db: %w", err)
	}

	return db, nil
}
