package cli

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/claude-watcher/internal/adapters/turso"
	sqlc "github.com/emiliopalmerini/claude-watcher/sqlc/generated"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data to JSON or CSV",
	Long: `Export session data for external analysis.

Examples:
  claude-watcher export sessions --format json --output sessions.json
  claude-watcher export sessions --format csv --output sessions.csv
  claude-watcher export sessions --experiment "baseline" --format json`,
}

var exportSessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "Export sessions data",
	RunE:  runExportSessions,
}

// Flags
var (
	exportFormat     string
	exportOutput     string
	exportExperiment string
	exportLimit      int
)

func init() {
	rootCmd.AddCommand(exportCmd)
	exportCmd.AddCommand(exportSessionsCmd)

	exportSessionsCmd.Flags().StringVarP(&exportFormat, "format", "f", "json", "Output format: json, csv")
	exportSessionsCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Output file (default: stdout)")
	exportSessionsCmd.Flags().StringVarP(&exportExperiment, "experiment", "e", "", "Filter by experiment name")
	exportSessionsCmd.Flags().IntVarP(&exportLimit, "limit", "n", 1000, "Maximum sessions to export")
}

type ExportSession struct {
	ID                     string  `json:"id"`
	ProjectID              string  `json:"project_id"`
	ExperimentID           string  `json:"experiment_id,omitempty"`
	Cwd                    string  `json:"cwd"`
	PermissionMode         string  `json:"permission_mode"`
	ExitReason             string  `json:"exit_reason"`
	StartedAt              string  `json:"started_at,omitempty"`
	EndedAt                string  `json:"ended_at,omitempty"`
	DurationSeconds        int64   `json:"duration_seconds,omitempty"`
	CreatedAt              string  `json:"created_at"`
	MessageCountUser       int64   `json:"message_count_user"`
	MessageCountAssistant  int64   `json:"message_count_assistant"`
	TurnCount              int64   `json:"turn_count"`
	TokenInput             int64   `json:"token_input"`
	TokenOutput            int64   `json:"token_output"`
	TokenCacheRead         int64   `json:"token_cache_read"`
	TokenCacheWrite        int64   `json:"token_cache_write"`
	CostEstimateUsd        float64 `json:"cost_estimate_usd,omitempty"`
	ErrorCount             int64   `json:"error_count"`
}

func runExportSessions(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db)

	var sessions []sqlc.Session

	if exportExperiment != "" {
		exp, err := queries.GetExperimentByName(ctx, exportExperiment)
		if err != nil {
			return fmt.Errorf("experiment %q not found", exportExperiment)
		}

		sessions, err = queries.ListSessionsByExperiment(ctx, sqlc.ListSessionsByExperimentParams{
			ExperimentID: toNullString(exp.ID),
			Limit:        int64(exportLimit),
		})
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
	} else {
		sessions, err = queries.ListSessions(ctx, int64(exportLimit))
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}
	}

	// Build export data with metrics
	exportData := make([]ExportSession, 0, len(sessions))
	for _, s := range sessions {
		es := ExportSession{
			ID:             s.ID,
			ProjectID:      s.ProjectID,
			Cwd:            s.Cwd,
			PermissionMode: s.PermissionMode,
			ExitReason:     s.ExitReason,
			CreatedAt:      s.CreatedAt,
		}

		if s.ExperimentID.Valid {
			es.ExperimentID = s.ExperimentID.String
		}
		if s.StartedAt.Valid {
			es.StartedAt = s.StartedAt.String
		}
		if s.EndedAt.Valid {
			es.EndedAt = s.EndedAt.String
		}
		if s.DurationSeconds.Valid {
			es.DurationSeconds = s.DurationSeconds.Int64
		}

		// Get metrics
		m, err := queries.GetSessionMetricsBySessionID(ctx, s.ID)
		if err == nil {
			es.MessageCountUser = m.MessageCountUser
			es.MessageCountAssistant = m.MessageCountAssistant
			es.TurnCount = m.TurnCount
			es.TokenInput = m.TokenInput
			es.TokenOutput = m.TokenOutput
			es.TokenCacheRead = m.TokenCacheRead
			es.TokenCacheWrite = m.TokenCacheWrite
			es.ErrorCount = m.ErrorCount
			if m.CostEstimateUsd.Valid {
				es.CostEstimateUsd = m.CostEstimateUsd.Float64
			}
		}

		exportData = append(exportData, es)
	}

	// Determine output
	var output *os.File
	if exportOutput != "" {
		output, err = os.Create(exportOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer output.Close()
	} else {
		output = os.Stdout
	}

	// Write output
	switch exportFormat {
	case "json":
		encoder := json.NewEncoder(output)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(exportData); err != nil {
			return fmt.Errorf("failed to encode JSON: %w", err)
		}
	case "csv":
		writer := csv.NewWriter(output)
		defer writer.Flush()

		// Header
		header := []string{
			"id", "project_id", "experiment_id", "cwd", "permission_mode", "exit_reason",
			"started_at", "ended_at", "duration_seconds", "created_at",
			"message_count_user", "message_count_assistant", "turn_count",
			"token_input", "token_output", "token_cache_read", "token_cache_write",
			"cost_estimate_usd", "error_count",
		}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}

		// Rows
		for _, es := range exportData {
			row := []string{
				es.ID, es.ProjectID, es.ExperimentID, es.Cwd, es.PermissionMode, es.ExitReason,
				es.StartedAt, es.EndedAt, fmt.Sprintf("%d", es.DurationSeconds), es.CreatedAt,
				fmt.Sprintf("%d", es.MessageCountUser), fmt.Sprintf("%d", es.MessageCountAssistant),
				fmt.Sprintf("%d", es.TurnCount), fmt.Sprintf("%d", es.TokenInput),
				fmt.Sprintf("%d", es.TokenOutput), fmt.Sprintf("%d", es.TokenCacheRead),
				fmt.Sprintf("%d", es.TokenCacheWrite), fmt.Sprintf("%.6f", es.CostEstimateUsd),
				fmt.Sprintf("%d", es.ErrorCount),
			}
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	default:
		return fmt.Errorf("unsupported format: %s (use json or csv)", exportFormat)
	}

	if exportOutput != "" {
		fmt.Fprintf(os.Stderr, "Exported %d sessions to %s\n", len(exportData), exportOutput)
	}

	return nil
}
