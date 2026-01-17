package turso

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func NewDB() (*sql.DB, error) {
	dbURL := os.Getenv("CLAUDE_WATCHER_DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("CLAUDE_WATCHER_DATABASE_URL environment variable is required")
	}

	authToken := os.Getenv("CLAUDE_WATCHER_AUTH_TOKEN")
	if authToken == "" {
		return nil, fmt.Errorf("CLAUDE_WATCHER_AUTH_TOKEN environment variable is required")
	}

	connStr := fmt.Sprintf("%s?authToken=%s", dbURL, authToken)
	db, err := sql.Open("libsql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
