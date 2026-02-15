package turso

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/tursodatabase/go-libsql"

	"github.com/emiliopalmerini/mclaude/internal/util"
)

// SyncMode indicates the database synchronization mode.
type SyncMode int

const (
	// SyncModeLocalOnly indicates no remote sync (credentials not available).
	SyncModeLocalOnly SyncMode = iota
	// SyncModeEnabled indicates sync with remote Turso database.
	SyncModeEnabled
)

// DB wraps the database connection and optional sync connector.
type DB struct {
	*sql.DB
	connector *libsql.Connector
	syncMode  SyncMode
	mu        sync.Mutex
}

// DBConfig holds configuration for database connection.
type DBConfig struct {
	// LocalPath is the path to the local database file.
	// If empty, defaults to XDG data directory.
	LocalPath string

	// RemoteURL is the Turso database URL (optional).
	// If empty, operates in local-only mode.
	RemoteURL string

	// AuthToken is the Turso authentication token (optional).
	// Required if RemoteURL is set.
	AuthToken string
}

// NewDB creates a database connection with configuration from environment variables.
// If MCLAUDE_DATABASE_URL and MCLAUDE_AUTH_TOKEN are set, enables sync.
// Otherwise, operates in local-only mode.
func NewDB() (*DB, error) {
	cfg := DBConfig{
		RemoteURL: os.Getenv("MCLAUDE_DATABASE_URL"),
		AuthToken: os.Getenv("MCLAUDE_AUTH_TOKEN"),
	}
	return NewDBWithConfig(cfg)
}

// NewDBWithConfig creates a database connection with explicit configuration.
func NewDBWithConfig(cfg DBConfig) (*DB, error) {
	// Determine local database path
	localPath := cfg.LocalPath
	if localPath == "" {
		dataDir, err := util.GetXDGDataDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get data directory: %w", err)
		}

		// Ensure directory exists
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create data directory: %w", err)
		}

		localPath = filepath.Join(dataDir, "mclaude.db")
	}

	// Determine sync mode
	syncEnabled := cfg.RemoteURL != "" && cfg.AuthToken != ""

	var db *sql.DB
	var connector *libsql.Connector
	var err error

	if syncEnabled {
		// Create embedded replica with sync
		connector, err = libsql.NewEmbeddedReplicaConnector(
			localPath,
			cfg.RemoteURL,
			libsql.WithAuthToken(cfg.AuthToken),
			libsql.WithReadYourWrites(true),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create embedded replica connector: %w", err)
		}

		db = sql.OpenDB(connector)
	} else {
		// Local-only mode using file path connection
		db, err = sql.Open("libsql", "file:"+localPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open local database: %w", err)
		}
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		if connector != nil {
			_ = connector.Close()
		}
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	syncMode := SyncModeLocalOnly
	if syncEnabled {
		syncMode = SyncModeEnabled
	}

	return &DB{
		DB:        db,
		connector: connector,
		syncMode:  syncMode,
	}, nil
}

// NewRemoteDB creates a direct remote connection to Turso (no embedded replica).
// Use this for server deployments where no local state is needed.
func NewRemoteDB(url, authToken string) (*sql.DB, error) {
	if url == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	if authToken == "" {
		return nil, fmt.Errorf("auth token is required")
	}

	db, err := sql.Open("libsql", url+"?authToken="+authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to open remote database: %w", err)
	}

	// Serialize all access through a single connection to avoid race conditions
	// in the go-libsql remote (HTTP/hrana) driver.
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping remote database: %w", err)
	}

	return db, nil
}

// Sync triggers an immediate sync with the remote database.
// Returns nil if sync is not enabled.
func (d *DB) Sync() error {
	if d.syncMode != SyncModeEnabled || d.connector == nil {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	_, err := d.connector.Sync()
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}
	return nil
}

// SyncMode returns the current synchronization mode.
func (d *DB) SyncMode() SyncMode {
	return d.syncMode
}

// IsSyncEnabled returns true if remote sync is enabled.
func (d *DB) IsSyncEnabled() bool {
	return d.syncMode == SyncModeEnabled
}

// Close closes the database connection and connector.
// Callers that need to push writes should call Sync() explicitly before Close().
func (d *DB) Close() error {
	var errs []error
	if err := d.DB.Close(); err != nil {
		errs = append(errs, err)
	}
	if d.connector != nil {
		if err := d.connector.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing database: %v", errs)
	}
	return nil
}
