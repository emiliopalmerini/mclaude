package turso_test

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/tursodatabase/go-libsql"

	"github.com/emiliopalmerini/mclaude/internal/migrate"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("libsql", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	ctx := context.Background()
	if err := migrate.RunAll(ctx, db); err != nil {
		_ = db.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	t.Cleanup(func() { _ = db.Close() })
	return db
}
