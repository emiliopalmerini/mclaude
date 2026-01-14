package database

import (
	"context"
	"database/sql"
	"strings"
	"time"

	_ "github.com/tursodatabase/go-libsql"
)

// Client wraps a SQL database connection with Turso-specific retry logic.
type Client struct {
	*sql.DB
}

// Options configures the database client behavior.
type Options struct {
	Ping bool
}

// New creates a new database client with default options (ping enabled).
func New(databaseURL, authToken string) (*Client, error) {
	return NewWithOptions(databaseURL, authToken, Options{Ping: true})
}

// NewNoPing creates a connection without an initial ping.
// Useful for hooks where latency matters and we'll discover failures on first query.
func NewNoPing(databaseURL, authToken string) (*Client, error) {
	return NewWithOptions(databaseURL, authToken, Options{Ping: false})
}

// NewWithOptions creates a database client with custom options.
func NewWithOptions(databaseURL, authToken string, opts Options) (*Client, error) {
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

	if opts.Ping {
		if err := db.Ping(); err != nil {
			return nil, err
		}
	}

	return &Client{DB: db}, nil
}

// IsStreamError checks if an error is a Turso "stream not found" error.
func IsStreamError(err error) bool {
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

		if !IsStreamError(err) || attempt == maxRetries {
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
