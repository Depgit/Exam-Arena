package database

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/exam-arena/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Migration is one embedded SQL file, identified by its numeric prefix.
type Migration struct {
	Version int
	Name    string
	SQL     string
}

// RunMigrations applies every embedded migration that is not yet recorded in
// schema_migrations. Each file runs inside its own transaction and is only
// recorded once it succeeds, so a failed migration is retried on next start.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migs, err := LoadMigrations()
	if err != nil {
		return err
	}

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range migs {
		var applied bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)`, m.Version,
		).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %d: %w", m.Version, err)
		}
		if applied {
			continue
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %d: %w", m.Version, err)
		}
		// Exec with no arguments uses the simple protocol, which allows the
		// multi-statement SQL that a schema file contains.
		if _, err := tx.Exec(ctx, m.SQL); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("apply migration %d (%s): %w", m.Version, m.Name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, m.Version); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %d: %w", m.Version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.Version, err)
		}
		slog.Info("migration applied", "version", m.Version, "name", m.Name)
	}

	return nil
}

// LoadMigrations reads every embedded *.sql file and returns them sorted by
// version. It fails on a missing numeric prefix, a duplicate version, or an
// empty file, so a broken migration set is caught at startup rather than
// silently recorded as applied.
func LoadMigrations() ([]Migration, error) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}

	seen := make(map[int]string)
	var migs []Migration
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".sql") {
			continue
		}

		prefix, _, _ := strings.Cut(name, "_")
		version, err := strconv.Atoi(prefix)
		if err != nil {
			return nil, fmt.Errorf("migration %q must start with a numeric version prefix", name)
		}
		if other, dup := seen[version]; dup {
			return nil, fmt.Errorf("migrations %q and %q share version %d", other, name, version)
		}
		seen[version] = name

		content, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", name, err)
		}
		if strings.TrimSpace(string(content)) == "" {
			return nil, fmt.Errorf("migration %q is empty", name)
		}

		migs = append(migs, Migration{Version: version, Name: name, SQL: string(content)})
	}

	if len(migs) == 0 {
		return nil, fmt.Errorf("no embedded migrations found")
	}

	sort.Slice(migs, func(i, j int) bool { return migs[i].Version < migs[j].Version })
	return migs, nil
}
