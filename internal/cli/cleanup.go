package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Delete sessions and their data",
	Long: `Delete sessions based on various filters.

Examples:
  mclaude cleanup --before 2024-01-01        # Delete sessions before date
  mclaude cleanup --project <id>             # Delete sessions for project
  mclaude cleanup --experiment <name>        # Delete sessions for experiment
  mclaude cleanup --session <id>             # Delete specific session
  mclaude cleanup --before 2024-01-01 --dry-run  # Preview what would be deleted`,
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

	var sessionsToDelete []domain.TranscriptPathInfo

	if cleanupSession != "" {
		session, err := app.SessionRepo.GetByID(ctx, cleanupSession)
		if err != nil {
			return fmt.Errorf("failed to get session: %w", err)
		}
		if session == nil {
			return fmt.Errorf("session %q not found", cleanupSession)
		}
		path := ""
		if session.TranscriptStoredPath != nil {
			path = *session.TranscriptStoredPath
		}
		sessionsToDelete = []domain.TranscriptPathInfo{{
			ID:             session.ID,
			TranscriptPath: path,
		}}
	} else if cleanupBefore != "" {
		beforeDate, err := time.Parse("2006-01-02", cleanupBefore)
		if err != nil {
			return fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", cleanupBefore)
		}
		beforeStr := beforeDate.Format(time.RFC3339)

		sessionsToDelete, err = app.SessionRepo.GetTranscriptPathsBefore(ctx, beforeStr)
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}
	} else if cleanupProject != "" {
		var err error
		sessionsToDelete, err = app.SessionRepo.GetTranscriptPathsByProject(ctx, cleanupProject)
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}
	} else if cleanupExperiment != "" {
		exp, err := getExperimentByName(ctx, app.ExperimentRepo, cleanupExperiment)
		if err != nil {
			return err
		}

		sessionsToDelete, err = app.SessionRepo.GetTranscriptPathsByExperiment(ctx, exp.ID)
		if err != nil {
			return fmt.Errorf("failed to get sessions: %w", err)
		}
	}

	if len(sessionsToDelete) == 0 {
		fmt.Println("No sessions to delete")
		return nil
	}

	if cleanupDryRun {
		fmt.Printf("Would delete %d session(s):\n", len(sessionsToDelete))
		for _, s := range sessionsToDelete {
			fmt.Printf("  - %s\n", s.ID)
		}
		return nil
	}

	deleted := 0
	for _, s := range sessionsToDelete {
		if err := app.SessionRepo.Delete(ctx, s.ID); err != nil {
			fmt.Printf("Warning: failed to delete session %s: %v\n", s.ID, err)
			continue
		}
		deleted++
	}

	fmt.Printf("Deleted %d session(s)\n", deleted)
	return nil
}
