package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/migrate"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [version]",
	Short: "Run database migrations",
	Long: `Run database migrations.

Without arguments, runs all pending migrations (up).
With a version number, migrates to that specific version (up or down as needed).

Examples:
  mclaude migrate      # Run all pending migrations
  mclaude migrate 5    # Migrate to version 5
  mclaude migrate 0    # Rollback all migrations`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMigrate,
}

func runMigrate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Connect to database
	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	// Ensure schema_migrations table exists
	if err := migrate.EnsureMigrationsTable(ctx, db.DB); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, dirty, err := migrate.GetCurrentVersion(ctx, db.DB)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state at version %d, manual intervention required", currentVersion)
	}

	// Load migrations
	allMigrations, err := migrate.LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	fmt.Printf("Current version: %d\n", currentVersion)

	var migrateErr error
	if len(args) == 0 {
		// Run all pending migrations (up)
		migrateErr = migrate.MigrateUp(ctx, db.DB, allMigrations, currentVersion)
	} else {
		// Migrate to specific version
		targetVersion, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid version number: %s", args[0])
		}

		if targetVersion > currentVersion {
			migrateErr = migrate.MigrateUpTo(ctx, db.DB, allMigrations, currentVersion, targetVersion)
		} else if targetVersion < currentVersion {
			migrateErr = migrate.MigrateDownTo(ctx, db.DB, allMigrations, currentVersion, targetVersion)
		} else {
			fmt.Println("Already at target version")
		}
	}

	// Sync schema changes to remote
	if err := db.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to sync migrations to remote: %v\n", err)
	}

	return migrateErr
}

