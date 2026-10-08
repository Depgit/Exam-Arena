package database

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

const (
	pgForeignKeyViolation = "23503"
	pgUniqueViolation     = "23505"
	pgInvalidTextRepr     = "22P02" // e.g. a malformed UUID literal
)

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

// IsForeignKeyViolation reports whether err is a Postgres FK failure, which
// means a referenced row (category, topic, user…) does not exist.
func IsForeignKeyViolation(err error) bool { return pgCode(err) == pgForeignKeyViolation }

// IsUniqueViolation reports whether err is a Postgres unique-constraint failure.
func IsUniqueViolation(err error) bool { return pgCode(err) == pgUniqueViolation }

// IsInvalidInput reports whether err was caused by a value Postgres could not
// parse for the column type, such as a string that is not a valid UUID.
func IsInvalidInput(err error) bool { return pgCode(err) == pgInvalidTextRepr }
