package cli

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/migrations"
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

type migration struct {
	version int
	name    string
	upSQL   string
	downSQL string
}

func runMigrate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Connect to database
	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Ensure schema_migrations table exists
	if err := ensureMigrationsTable(ctx, db.DB); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get current version
	currentVersion, dirty, err := getCurrentVersion(ctx, db.DB)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state at version %d, manual intervention required", currentVersion)
	}

	// Load migrations
	allMigrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	fmt.Printf("Current version: %d\n", currentVersion)

	var migrateErr error
	if len(args) == 0 {
		// Run all pending migrations (up)
		migrateErr = migrateUp(ctx, db.DB, allMigrations, currentVersion)
	} else {
		// Migrate to specific version
		targetVersion, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid version number: %s", args[0])
		}

		if targetVersion > currentVersion {
			migrateErr = migrateUpTo(ctx, db.DB, allMigrations, currentVersion, targetVersion)
		} else if targetVersion < currentVersion {
			migrateErr = migrateDownTo(ctx, db.DB, allMigrations, currentVersion, targetVersion)
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

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	// Check if table exists and has the dirty column
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pragma_table_info('schema_migrations') WHERE name = 'dirty'
	`).Scan(&count)

	if err != nil {
		// Table might not exist, create it
		_, err = db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INTEGER PRIMARY KEY,
				dirty INTEGER NOT NULL DEFAULT 0
			)
		`)
		return err
	}

	if count == 0 {
		// Table exists but doesn't have dirty column - drop and recreate
		_, err = db.ExecContext(ctx, `DROP TABLE IF EXISTS schema_migrations`)
		if err != nil {
			return err
		}
		_, err = db.ExecContext(ctx, `
			CREATE TABLE schema_migrations (
				version INTEGER PRIMARY KEY,
				dirty INTEGER NOT NULL DEFAULT 0
			)
		`)
		return err
	}

	return nil
}

func getCurrentVersion(ctx context.Context, db *sql.DB) (int, bool, error) {
	var version int
	var dirty int

	err := db.QueryRowContext(ctx, `SELECT version, dirty FROM schema_migrations ORDER BY version DESC LIMIT 1`).Scan(&version, &dirty)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}

	return version, dirty == 1, nil
}

func setVersion(ctx context.Context, db *sql.DB, version int, dirty bool) error {
	dirtyInt := 0
	if dirty {
		dirtyInt = 1
	}

	// Delete all existing versions and insert new one
	_, err := db.ExecContext(ctx, `DELETE FROM schema_migrations`)
	if err != nil {
		return err
	}

	if version > 0 {
		_, err = db.ExecContext(ctx, `INSERT INTO schema_migrations (version, dirty) VALUES (?, ?)`, version, dirtyInt)
	}
	return err
}

func loadMigrations() ([]migration, error) {
	var result []migration

	// Regex to match migration files
	upPattern := regexp.MustCompile(`^(\d+)_(.+)\.up\.sql$`)

	err := fs.WalkDir(migrations.FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		matches := upPattern.FindStringSubmatch(filepath.Base(path))
		if matches == nil {
			return nil
		}

		version, _ := strconv.Atoi(matches[1])
		name := matches[2]

		// Read up SQL
		upSQL, err := fs.ReadFile(migrations.FS, path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Read down SQL
		downPath := fmt.Sprintf("%03d_%s.down.sql", version, name)
		downSQL, err := fs.ReadFile(migrations.FS, downPath)
		if err != nil {
			// Down migration is optional
			downSQL = nil
		}

		result = append(result, migration{
			version: version,
			name:    name,
			upSQL:   string(upSQL),
			downSQL: string(downSQL),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by version
	sort.Slice(result, func(i, j int) bool {
		return result[i].version < result[j].version
	})

	return result, nil
}

func migrateUp(ctx context.Context, db *sql.DB, allMigrations []migration, currentVersion int) error {
	fmt.Println("Running all pending migrations...")

	count := 0
	for _, m := range allMigrations {
		if m.version <= currentVersion {
			continue
		}

		if err := runMigration(ctx, db, m, true); err != nil {
			return err
		}
		count++
	}

	if count == 0 {
		fmt.Println("No migrations to run")
	} else {
		newVersion, _, _ := getCurrentVersion(ctx, db)
		fmt.Printf("Migrated to version %d (%d migrations applied)\n", newVersion, count)
	}

	return nil
}

func migrateUpTo(ctx context.Context, db *sql.DB, allMigrations []migration, currentVersion, targetVersion int) error {
	fmt.Printf("Migrating up to version %d...\n", targetVersion)

	for _, m := range allMigrations {
		if m.version <= currentVersion {
			continue
		}
		if m.version > targetVersion {
			break
		}

		if err := runMigration(ctx, db, m, true); err != nil {
			return err
		}
	}

	fmt.Printf("Migrated to version %d\n", targetVersion)
	return nil
}

func migrateDownTo(ctx context.Context, db *sql.DB, allMigrations []migration, currentVersion, targetVersion int) error {
	fmt.Printf("Migrating down to version %d...\n", targetVersion)

	// Reverse order for down migrations
	for i := len(allMigrations) - 1; i >= 0; i-- {
		m := allMigrations[i]
		if m.version > currentVersion {
			continue
		}
		if m.version <= targetVersion {
			break
		}

		if m.downSQL == "" {
			return fmt.Errorf("no down migration for version %d", m.version)
		}

		if err := runMigration(ctx, db, m, false); err != nil {
			return err
		}
	}

	fmt.Printf("Migrated to version %d\n", targetVersion)
	return nil
}

func runMigration(ctx context.Context, db *sql.DB, m migration, up bool) error {
	direction := "up"
	sqlContent := m.upSQL
	if !up {
		direction = "down"
		sqlContent = m.downSQL
	}

	fmt.Printf("  %s %d_%s...\n", direction, m.version, m.name)

	// Set dirty flag before running
	targetVersion := m.version
	if !up {
		targetVersion = m.version - 1
	}
	if err := setVersion(ctx, db, m.version, true); err != nil {
		return fmt.Errorf("failed to set dirty flag: %w", err)
	}

	// Split SQL by semicolons and execute each statement
	statements := splitSQL(sqlContent)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute migration %d %s: %w\nSQL: %s", m.version, direction, err, stmt)
		}
	}

	// Clear dirty flag
	if err := setVersion(ctx, db, targetVersion, false); err != nil {
		return fmt.Errorf("failed to clear dirty flag: %w", err)
	}

	return nil
}

func splitSQL(sql string) []string {
	// Simple split by semicolon - doesn't handle semicolons in strings
	// but that's fine for our migrations
	return strings.Split(sql, ";")
}

func init() {
	// Silence unused import warning
	_ = os.Stderr
}

// RunMigrations runs all pending migrations on the provided database.
// Exported for use in tests.
func RunMigrations(ctx context.Context, db *sql.DB) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	currentVersion, dirty, err := getCurrentVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state at version %d", currentVersion)
	}

	allMigrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	for _, m := range allMigrations {
		if m.version <= currentVersion {
			continue
		}

		if err := runMigration(ctx, db, m, true); err != nil {
			return err
		}
	}

	return nil
}
