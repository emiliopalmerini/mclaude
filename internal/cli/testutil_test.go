package cli

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	_ "github.com/tursodatabase/go-libsql"

	"github.com/emiliopalmerini/mclaude/internal/migrate"
)

// testDB creates an in-memory SQLite database with all migrations applied.
// This is fast and suitable for most unit/integration tests.
func testDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	db, err := sql.Open("libsql", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Run migrations
	ctx := context.Background()
	if err := migrate.RunAll(ctx, db); err != nil {
		_ = db.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
	}

	return db, cleanup
}

// testTursoDB creates a Turso (libsql-server) container for full integration testing.
// This is slower but tests against the real Turso server.
func testTursoDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	ctx := context.Background()

	// Start libsql-server container
	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/tursodatabase/libsql-server:latest",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/health").WithPort("8080/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Turso container: %v", err)
	}

	// Get the mapped port
	mappedPort, err := container.MappedPort(ctx, "8080")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("Failed to get container host: %v", err)
	}

	// Connect to the database
	url := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	db, err := sql.Open("libsql", url)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("Failed to connect to Turso: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		t.Fatalf("Failed to ping Turso: %v", err)
	}

	// Run migrations
	if err := migrate.RunAll(ctx, db); err != nil {
		_ = db.Close()
		_ = container.Terminate(ctx)
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
		_ = container.Terminate(ctx)
	}

	return db, cleanup
}
