package database

import (
	"context"
	"database/sql"
	"strings"
	"time"

	_ "github.com/tursodatabase/go-libsql"
)

func NewTurso(databaseURL, authToken string) (*sql.DB, error) {
	return NewTursoWithOptions(databaseURL, authToken, true)
}

// NewTursoNoPing creates a connection without an initial ping.
// Useful for hooks where latency matters and we'll discover failures on first query.
func NewTursoNoPing(databaseURL, authToken string) (*sql.DB, error) {
	return NewTursoWithOptions(databaseURL, authToken, false)
}

func NewTursoWithOptions(databaseURL, authToken string, ping bool) (*sql.DB, error) {
	connStr := databaseURL + "?authToken=" + authToken
	db, err := sql.Open("libsql", connStr)
	if err != nil {
		return nil, err
	}

	// Configure connection pool for Turso's Hrana protocol.
	// Use minimal idle connections since Turso aggressively closes
	// idle streams, causing "stream not found" errors on stale connections.
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(0) // Disable idle connections to force fresh connections
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(0) // Don't keep idle connections

	if ping {
		if err := db.Ping(); err != nil {
			return nil, err
		}
	}

	return db, nil
}

// IsTursoStreamError checks if an error is a Turso "stream not found" error.
func IsTursoStreamError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "stream not found")
}

// WithRetry executes a function with retry logic for Turso stream errors.
// It retries up to maxRetries times when encountering "stream not found" errors.
func WithRetry[T any](ctx context.Context, maxRetries int, fn func() (T, error)) (T, error) {
	var result T
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err = fn()
		if err == nil {
			return result, nil
		}

		if !IsTursoStreamError(err) || attempt == maxRetries {
			return result, err
		}

		// Brief pause before retry to allow connection pool to refresh
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}

	return result, err
}
