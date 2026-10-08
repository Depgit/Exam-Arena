package database

import (
	"strings"
	"testing"
)

// The original code had a commented-out go:embed directive, so the schema was
// an empty string and a fresh database ended up with no tables. This test
// guards the embedding itself: it fails if the schema stops being embedded.
func TestLoadMigrationsEmbedsSchema(t *testing.T) {
	migs, err := LoadMigrations()
	if err != nil {
		t.Fatalf("LoadMigrations: %v", err)
	}
	if migs[0].Version != 1 {
		t.Fatalf("first migration version = %d, want 1", migs[0].Version)
	}
	for i := 1; i < len(migs); i++ {
		if migs[i].Version <= migs[i-1].Version {
			t.Fatalf("migrations not sorted: %d after %d", migs[i].Version, migs[i-1].Version)
		}
	}
	for _, want := range []string{
		"CREATE TABLE IF NOT EXISTS users",
		"CREATE TABLE IF NOT EXISTS exam_categories",
		"CREATE TABLE IF NOT EXISTS topics",
		"CREATE TABLE IF NOT EXISTS questions",
		"CREATE TABLE IF NOT EXISTS matchmaking_queue",
	} {
		if !strings.Contains(migs[0].SQL, want) {
			t.Errorf("initial schema is missing %q", want)
		}
	}
}
