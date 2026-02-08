package turso_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/tursodatabase/go-libsql"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
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
		db.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	t.Cleanup(func() { db.Close() })
	return db
}

func writeTempTranscript(t *testing.T, content string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "transcript.jsonl")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write temp transcript: %v", err)
	}
	return path
}

func TestTranscriptRepository_StoreAndGet(t *testing.T) {
	db := testDB(t)
	repo := turso.NewTranscriptRepository(db)
	ctx := context.Background()

	content := `{"type":"message","role":"user","content":"hello"}
{"type":"message","role":"assistant","content":"hi"}
`
	path := writeTempTranscript(t, content)

	storedPath, err := repo.Store(ctx, "session-1", path)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}
	if storedPath != "db" {
		t.Errorf("expected storedPath = %q, got %q", "db", storedPath)
	}

	data, err := repo.Get(ctx, "session-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if string(data) != content {
		t.Errorf("roundtrip mismatch:\ngot:  %q\nwant: %q", string(data), content)
	}
}

func TestTranscriptRepository_Exists(t *testing.T) {
	db := testDB(t)
	repo := turso.NewTranscriptRepository(db)
	ctx := context.Background()

	exists, err := repo.Exists(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("expected Exists to return false for nonexistent session")
	}

	path := writeTempTranscript(t, "test data")
	if _, err := repo.Store(ctx, "session-2", path); err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	exists, err = repo.Exists(ctx, "session-2")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("expected Exists to return true after Store")
	}
}

func TestTranscriptRepository_Delete(t *testing.T) {
	db := testDB(t)
	repo := turso.NewTranscriptRepository(db)
	ctx := context.Background()

	path := writeTempTranscript(t, "delete me")
	if _, err := repo.Store(ctx, "session-3", path); err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	if err := repo.Delete(ctx, "session-3"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	exists, err := repo.Exists(ctx, "session-3")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Error("expected Exists to return false after Delete")
	}
}

func TestTranscriptRepository_GetNonexistent(t *testing.T) {
	db := testDB(t)
	repo := turso.NewTranscriptRepository(db)
	ctx := context.Background()

	_, err := repo.Get(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error for Get of nonexistent session")
	}
}
