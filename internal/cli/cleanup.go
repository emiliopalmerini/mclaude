package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/claude-watcher/internal/adapters/storage"
	"github.com/emiliopalmerini/claude-watcher/internal/adapters/turso"
	sqlc "github.com/emiliopalmerini/claude-watcher/sqlc/generated"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Delete sessions and their data",
	Long: `Delete sessions based on various filters.

Examples:
  claude-watcher cleanup --before 2024-01-01        # Delete sessions before date
  claude-watcher cleanup --project <id>             # Delete sessions for project
  claude-watcher cleanup --experiment <name>        # Delete sessions for experiment
  claude-watcher cleanup --session <id>             # Delete specific session
  claude-watcher cleanup --before 2024-01-01 --dry-run  # Preview what would be deleted`,
	RunE: runCleanup,
}

// Flags
var (
	cleanupBefore     string
	cleanupProject    string
	cleanupExperiment string
	cleanupSession    string
	cleanupDryRun     bool
)

func init() {
	rootCmd.AddCommand(cleanupCmd)

	cleanupCmd.Flags().StringVar(&cleanupBefore, "before", "", "Delete sessions before date (YYYY-MM-DD)")
	cleanupCmd.Flags().StringVar(&cleanupProject, "project", "", "Delete sessions for project ID")
	cleanupCmd.Flags().StringVar(&cleanupExperiment, "experiment", "", "Delete sessions for experiment name")
	cleanupCmd.Flags().StringVar(&cleanupSession, "session", "", "Delete specific session ID")
	cleanupCmd.Flags().BoolVar(&cleanupDryRun, "dry-run", false, "Preview what would be deleted")
}

func runCleanup(cmd *cobra.Command, args []string) error {
	if cleanupBefore == "" && cleanupProject == "" && cleanupExperiment == "" && cleanupSession == "" {
		return fmt.Errorf("at least one filter is required: --before, --project, --experiment, or --session")
	}

	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db)

	// Initialize transcript storage for cleanup
	transcriptStorage, err := storage.NewTranscriptStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize transcript storage: %w", err)
	}

	var sessionsToDelete []sessionInfo

	if cleanupSession != "" {
		// Delete specific session
		session, err := queries.GetSessionByID(ctx, cleanupSession)
		if err != nil {
			return fmt.Errorf("session %q not found", cleanupSession)
		}
		sessionsToDelete = []sessionInfo{{
			id:             session.ID,
			transcriptPath: session.TranscriptStoredPath.String,
		}}
	} else if cleanupBefore != "" {
		// Parse date
		beforeDate, err := time.Parse("2006-01-02", cleanupBefore)
		if err != nil {
			return fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", cleanupBefore)
		}
		beforeStr := beforeDate.Format(time.RFC3339)

		// Get sessions to delete
		paths, err := queries.GetSessionTranscriptPathsBefore(ctx, beforeStr)
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}
		for _, p := range paths {
			sessionsToDelete = append(sessionsToDelete, sessionInfo{
				id:             p.ID,
				transcriptPath: p.TranscriptStoredPath.String,
			})
		}
	} else if cleanupProject != "" {
		paths, err := queries.GetSessionTranscriptPathsByProject(ctx, cleanupProject)
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}
		for _, p := range paths {
			sessionsToDelete = append(sessionsToDelete, sessionInfo{
				id:             p.ID,
				transcriptPath: p.TranscriptStoredPath.String,
			})
		}
	} else if cleanupExperiment != "" {
		// Get experiment ID
		exp, err := queries.GetExperimentByName(ctx, cleanupExperiment)
		if err != nil {
			return fmt.Errorf("experiment %q not found", cleanupExperiment)
		}

		paths, err := queries.GetSessionTranscriptPathsByExperiment(ctx, toNullString(exp.ID))
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}
		for _, p := range paths {
			sessionsToDelete = append(sessionsToDelete, sessionInfo{
				id:             p.ID,
				transcriptPath: p.TranscriptStoredPath.String,
			})
		}
	}

	if len(sessionsToDelete) == 0 {
		fmt.Println("No sessions to delete")
		return nil
	}

	if cleanupDryRun {
		fmt.Printf("Would delete %d session(s):\n", len(sessionsToDelete))
		for _, s := range sessionsToDelete {
			fmt.Printf("  - %s\n", s.id)
		}
		return nil
	}

	// Delete sessions and transcripts
	deleted := 0
	for _, s := range sessionsToDelete {
		// Delete transcript file
		if s.transcriptPath != "" {
			if err := transcriptStorage.Delete(ctx, s.id); err != nil {
				fmt.Printf("Warning: failed to delete transcript for %s: %v\n", s.id, err)
			}
		}

		// Delete session (cascades to metrics, tools, files, commands)
		if err := queries.DeleteSession(ctx, s.id); err != nil {
			fmt.Printf("Warning: failed to delete session %s: %v\n", s.id, err)
			continue
		}
		deleted++
	}

	fmt.Printf("Deleted %d session(s)\n", deleted)
	return nil
}

type sessionInfo struct {
	id             string
	transcriptPath string
}
