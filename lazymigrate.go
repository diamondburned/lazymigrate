// package lazymigrate allows for simple SQLite migrations by using a magic
// comment to delimit different versions of the schema.
package lazymigrate

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Delimiter is the default delimiter for the schema string.
// It is intentionally long and ugly to avoid collisions.
const Delimiter = "--------------------------------- NEW VERSION ---------------------------------"

// Schema wraps a SQLite schema string. A schema string is a series of SQL
// statements that create and modify tables. The schema string is delimited by
// a configurable magic comment. The magic comment must be on its own line
// before and after the schema string. It must not appear at the start or end
// of the schema string.
type Schema struct {
	schema string
	magic  string
}

// NewSchema returns a new Schema with the given schema string. The schema
// string is delimited by the default magic comment [Delimiter].
func NewSchema(schema string) *Schema {
	return NewSchemaWithMagic(schema, Delimiter)
}

// NewSchemaWithMagic returns a new Schema with the given schema string and
// magic comment.
func NewSchemaWithMagic(schema, magic string) *Schema {
	return &Schema{
		schema: schema,
		magic:  magic,
	}
}

// Versions returns the versions of the schema.
func (s *Schema) Versions() []string {
	return strings.Split(s.schema, "\n"+s.magic+"\n")
}

// Migrate migrates the database at the given source to the latest migrations.
// It uses the user_version pragma. Note that the function does not set any
// pragma values except for user_version. If you need to set other pragmas,
// you must do so yourself.
//
// The migrations are all done in a single transaction. If any migration fails,
// the transaction is rolled back and the error is returned.
func (s *Schema) Migrate(ctx context.Context, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	var v int
	if err := tx.QueryRowContext(ctx, "PRAGMA user_version").Scan(&v); err != nil {
		return fmt.Errorf("cannot get PRAGMA user_version: %w", err)
	}

	versions := s.Versions()
	if v >= len(versions) {
		return nil
	}

	for i := v; i < len(versions); i++ {
		_, err := tx.ExecContext(ctx, versions[i])
		if err != nil {
			return fmt.Errorf("cannot apply migration %d (from 0th): %w", i, err)
		}
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintln("PRAGMA user_version =", len(versions))); err != nil {
		return fmt.Errorf("cannot set PRAGMA user_version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("cannot commit new migrations: %w", err)
	}

	return nil
}

// Migrate migrates the database at the given source to the latest migrations.
// It is a convenience function around [NewSchema] and [Schema.Migrate].
func Migrate(ctx context.Context, db *sql.DB, schema string) error {
	return NewSchema(schema).Migrate(ctx, db)
}
