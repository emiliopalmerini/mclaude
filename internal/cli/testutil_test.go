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
		db.Close()
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Run migrations
	ctx := context.Background()
	if err := RunMigrations(ctx, db); err != nil {
		db.Close()
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

// tursoContainer holds the Turso container and connection info
type tursoContainer struct {
	container testcontainers.Container
	db        *sql.DB
	url       string
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
		container.Terminate(ctx)
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to get container host: %v", err)
	}

	// Connect to the database
	url := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	db, err := sql.Open("libsql", url)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to connect to Turso: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		container.Terminate(ctx)
		t.Fatalf("Failed to ping Turso: %v", err)
	}

	// Run migrations
	if err := RunMigrations(ctx, db); err != nil {
		db.Close()
		container.Terminate(ctx)
		t.Fatalf("Failed to run migrations: %v", err)
	}

	cleanup := func() {
		db.Close()
		container.Terminate(ctx)
	}

	return db, cleanup
}

// testDBType specifies which database backend to use for tests
type testDBType int

const (
	// DBTypeMemory uses in-memory SQLite (fast, for most tests)
	DBTypeMemory testDBType = iota
	// DBTypeTurso uses Turso container (slower, for full integration)
	DBTypeTurso
)

// getTestDB returns a test database based on the specified type.
// Use DBTypeMemory for fast tests, DBTypeTurso for full integration.
func getTestDB(t *testing.T, dbType testDBType) (*sql.DB, func()) {
	t.Helper()

	switch dbType {
	case DBTypeTurso:
		return testTursoDB(t)
	default:
		return testDB(t)
	}
}
