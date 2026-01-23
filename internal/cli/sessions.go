package cli

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/util"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Manage sessions",
	Long:  `List and manage recorded sessions.`,
}

var sessionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sessions",
	Long: `List recorded sessions with optional filters.

Examples:
  mclaude sessions list                     # Last 10 sessions
  mclaude sessions list --last 20           # Last 20 sessions
  mclaude sessions list --experiment "exp"  # Sessions for experiment
  mclaude sessions list --project <id>      # Sessions for project`,
	RunE: runSessionsList,
}

// Flags
var (
	sessionsLast       int
	sessionsExperiment string
	sessionsProject    string
)

func init() {
	rootCmd.AddCommand(sessionsCmd)
	sessionsCmd.AddCommand(sessionsListCmd)

	sessionsListCmd.Flags().IntVarP(&sessionsLast, "last", "n", 10, "Number of sessions to show")
	sessionsListCmd.Flags().StringVarP(&sessionsExperiment, "experiment", "e", "", "Filter by experiment name")
	sessionsListCmd.Flags().StringVar(&sessionsProject, "project", "", "Filter by project ID")
}

func runSessionsList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db.DB)

	var sessions []sqlc.Session

	if sessionsExperiment != "" {
		// Get experiment ID by name
		exp, err := queries.GetExperimentByName(ctx, sessionsExperiment)
		if err != nil {
			return fmt.Errorf("experiment %q not found", sessionsExperiment)
		}

		sessions, err = queries.ListSessionsByExperiment(ctx, sqlc.ListSessionsByExperimentParams{
			ExperimentID: sql.NullString{String: exp.ID, Valid: true},
			Limit:        int64(sessionsLast),
		})
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
	} else if sessionsProject != "" {
		sessions, err = queries.ListSessionsByProject(ctx, sqlc.ListSessionsByProjectParams{
			ProjectID: sessionsProject,
			Limit:     int64(sessionsLast),
		})
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
	} else {
		sessions, err = queries.ListSessions(ctx, int64(sessionsLast))
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions found")
		return nil
	}

	// Get metrics for each session
	type sessionWithMetrics struct {
		session sqlc.Session
		metrics *sqlc.SessionMetric
	}

	sessionsWithMetrics := make([]sessionWithMetrics, len(sessions))
	for i, s := range sessions {
		sessionsWithMetrics[i].session = s
		m, err := queries.GetSessionMetricsBySessionID(ctx, s.ID)
		if err == nil {
			sessionsWithMetrics[i].metrics = &m
		}
	}

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tDATE\tTURNS\tTOKENS\tCOST\tREASON")
	fmt.Fprintln(w, "--\t----\t-----\t------\t----\t------")

	for _, sm := range sessionsWithMetrics {
		s := sm.session
		m := sm.metrics

		// Format session ID (first 12 chars)
		id := s.ID
		if len(id) > 12 {
			id = id[:12]
		}

		// Format date
		date := util.FormatDateTime(s.CreatedAt)

		// Format metrics
		turns := "-"
		tokens := "-"
		cost := "-"
		if m != nil {
			turns = fmt.Sprintf("%d", m.TurnCount)
			totalTokens := m.TokenInput + m.TokenOutput
			tokens = util.FormatNumber(totalTokens)
			if m.CostEstimateUsd.Valid {
				cost = fmt.Sprintf("$%.4f", m.CostEstimateUsd.Float64)
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", id, date, turns, tokens, cost, s.ExitReason)
	}

	w.Flush()

	fmt.Printf("\nShowing %d session(s)\n", len(sessions))
	return nil
}
