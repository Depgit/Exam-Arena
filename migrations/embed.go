// Package migrations embeds the SQL schema files so the compiled binary is
// self-contained and never depends on the working directory at runtime.
package migrations

import "embed"

// FS holds every *.sql file in this directory.
//
// database.RunMigrations applies them in ascending version order, where the
// version is the numeric prefix of the filename (001_initial_schema.sql → 1).
// Add a new migration by dropping a NNN_description.sql file here.
//
//go:embed *.sql
var FS embed.FS
