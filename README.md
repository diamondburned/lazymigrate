# lazymigrate

Package lazymigrate allows for simple SQLite migrations by using a magic
comment to delimit different versions of the schema.

This package is meant to work with sqlc, which requires you to keep a separate
`schema.sql` file for your database schema. This package allows you to keep
multiple versions of the schema in the same file, and then use a magic comment
to delimit the different versions.

## Usage

Example SQL file:

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

--------------------------------- NEW VERSION ---------------------------------

ALTER TABLE users ADD COLUMN email TEXT;
```

```go
//go:embed schema.sql
var schema string

func main() {
    db, _ := sql.Open("sqlite3", "file::memory:")
    defer db.Close()

    migration := lazymigrate.NewSchema(schema)
    if err := migration.Migrate(context.TODO(), db); err != nil {
        log.Fatal(err)
    }
}
```
