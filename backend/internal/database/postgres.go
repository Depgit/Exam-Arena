package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresSQLDB(ctx context.Context, connStr string, maxConns, minConns int32) (*sql.DB, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("open postgres db: %w", err)
	}

	db.SetMaxOpenConns(int(maxConns))
	db.SetMaxIdleConns(int(minConns))

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping postgres db: %w", err)
	}

	return db, nil
}

func NewPostgres(ctx context.Context, connStr string, maxConns, minConns int32) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.MaxConns = maxConns
	cfg.MinConns = minConns

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return pool, nil
}
