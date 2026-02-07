package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/emiliopalmerini/mclaude/migrations"
)

// Migration represents a single database migration with up and down SQL.
type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

// EnsureMigrationsTable creates the schema_migrations table if it doesn't exist.
func EnsureMigrationsTable(ctx context.Context, db *sql.DB) error {
	var count int
	err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pragma_table_info('schema_migrations') WHERE name = 'dirty'
	`).Scan(&count)

	if err != nil {
		_, err = db.ExecContext(ctx, `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				version INTEGER PRIMARY KEY,
				dirty INTEGER NOT NULL DEFAULT 0
			)
		`)
		return err
	}

	if count == 0 {
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

// GetCurrentVersion returns the current migration version and dirty state.
func GetCurrentVersion(ctx context.Context, db *sql.DB) (int, bool, error) {
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

// SetVersion sets the migration version and dirty state.
func SetVersion(ctx context.Context, db *sql.DB, version int, dirty bool) error {
	dirtyInt := 0
	if dirty {
		dirtyInt = 1
	}

	_, err := db.ExecContext(ctx, `DELETE FROM schema_migrations`)
	if err != nil {
		return err
	}

	if version > 0 {
		_, err = db.ExecContext(ctx, `INSERT INTO schema_migrations (version, dirty) VALUES (?, ?)`, version, dirtyInt)
	}
	return err
}

// LoadMigrations reads all embedded migration files and returns them sorted by version.
func LoadMigrations() ([]Migration, error) {
	var result []Migration

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

		upSQL, err := fs.ReadFile(migrations.FS, path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		downPath := fmt.Sprintf("%03d_%s.down.sql", version, name)
		downSQL, err := fs.ReadFile(migrations.FS, downPath)
		if err != nil {
			downSQL = nil
		}

		result = append(result, Migration{
			Version: version,
			Name:    name,
			UpSQL:   string(upSQL),
			DownSQL: string(downSQL),
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result, nil
}

// RunMigration executes a single migration (up or down).
func RunMigration(ctx context.Context, db *sql.DB, m Migration, up bool) error {
	direction := "up"
	sqlContent := m.UpSQL
	if !up {
		direction = "down"
		sqlContent = m.DownSQL
	}

	fmt.Printf("  %s %d_%s...\n", direction, m.Version, m.Name)

	targetVersion := m.Version
	if !up {
		targetVersion = m.Version - 1
	}
	if err := SetVersion(ctx, db, m.Version, true); err != nil {
		return fmt.Errorf("failed to set dirty flag: %w", err)
	}

	statements := SplitSQL(sqlContent)
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute migration %d %s: %w\nSQL: %s", m.Version, direction, err, stmt)
		}
	}

	if err := SetVersion(ctx, db, targetVersion, false); err != nil {
		return fmt.Errorf("failed to clear dirty flag: %w", err)
	}

	return nil
}

// SplitSQL splits a SQL string by semicolons.
func SplitSQL(sql string) []string {
	return strings.Split(sql, ";")
}

// MigrateUp runs all pending up migrations.
func MigrateUp(ctx context.Context, db *sql.DB, allMigrations []Migration, currentVersion int) error {
	fmt.Println("Running all pending migrations...")

	count := 0
	for _, m := range allMigrations {
		if m.Version <= currentVersion {
			continue
		}

		if err := RunMigration(ctx, db, m, true); err != nil {
			return err
		}
		count++
	}

	if count == 0 {
		fmt.Println("No migrations to run")
	} else {
		newVersion, _, _ := GetCurrentVersion(ctx, db)
		fmt.Printf("Migrated to version %d (%d migrations applied)\n", newVersion, count)
	}

	return nil
}

// MigrateUpTo runs up migrations to a specific version.
func MigrateUpTo(ctx context.Context, db *sql.DB, allMigrations []Migration, currentVersion, targetVersion int) error {
	fmt.Printf("Migrating up to version %d...\n", targetVersion)

	for _, m := range allMigrations {
		if m.Version <= currentVersion {
			continue
		}
		if m.Version > targetVersion {
			break
		}

		if err := RunMigration(ctx, db, m, true); err != nil {
			return err
		}
	}

	fmt.Printf("Migrated to version %d\n", targetVersion)
	return nil
}

// MigrateDownTo runs down migrations to a specific version.
func MigrateDownTo(ctx context.Context, db *sql.DB, allMigrations []Migration, currentVersion, targetVersion int) error {
	fmt.Printf("Migrating down to version %d...\n", targetVersion)

	for i := len(allMigrations) - 1; i >= 0; i-- {
		m := allMigrations[i]
		if m.Version > currentVersion {
			continue
		}
		if m.Version <= targetVersion {
			break
		}

		if m.DownSQL == "" {
			return fmt.Errorf("no down migration for version %d", m.Version)
		}

		if err := RunMigration(ctx, db, m, false); err != nil {
			return err
		}
	}

	fmt.Printf("Migrated to version %d\n", targetVersion)
	return nil
}

// RunAll runs all pending migrations on the provided database.
func RunAll(ctx context.Context, db *sql.DB) error {
	if err := EnsureMigrationsTable(ctx, db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	currentVersion, dirty, err := GetCurrentVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		return fmt.Errorf("database is in dirty state at version %d", currentVersion)
	}

	allMigrations, err := LoadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	for _, m := range allMigrations {
		if m.Version <= currentVersion {
			continue
		}

		if err := RunMigration(ctx, db, m, true); err != nil {
			return err
		}
	}

	return nil
}
