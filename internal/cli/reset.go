package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the database (drop all tables)",
	Long: `Reset the database by dropping all tables.

WARNING: This will delete ALL data. Use with caution.`,
	RunE: runReset,
}

func init() {
	rootCmd.AddCommand(resetCmd)
}

func runReset(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Get all table names
	rows, err := db.QueryContext(ctx, `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'`)
	if err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, name)
	}

	if len(tables) == 0 {
		fmt.Println("No tables to drop")
		return nil
	}

	fmt.Printf("Dropping %d tables...\n", len(tables))

	// Disable foreign key checks temporarily
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("failed to disable foreign keys: %w", err)
	}

	// Drop each table
	for _, table := range tables {
		fmt.Printf("  Dropping %s...\n", table)
		if _, err := db.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
	}

	// Re-enable foreign key checks
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	fmt.Println("Database reset complete")
	return nil
}
