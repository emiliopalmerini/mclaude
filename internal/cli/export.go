package cli

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data to JSON or CSV",
	Long: `Export session data for external analysis.

Examples:
  mclaude export sessions --format json --output sessions.json
  mclaude export sessions --format csv --output sessions.csv
  mclaude export sessions --experiment "baseline" --format json`,
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
	ID                    string  `json:"id"`
	ProjectID             string  `json:"project_id"`
	ExperimentID          string  `json:"experiment_id,omitempty"`
	Cwd                   string  `json:"cwd"`
	PermissionMode        string  `json:"permission_mode"`
	ExitReason            string  `json:"exit_reason"`
	StartedAt             string  `json:"started_at,omitempty"`
	EndedAt               string  `json:"ended_at,omitempty"`
	DurationSeconds       int64   `json:"duration_seconds,omitempty"`
	CreatedAt             string  `json:"created_at"`
	MessageCountUser      int64   `json:"message_count_user"`
	MessageCountAssistant int64   `json:"message_count_assistant"`
	TurnCount             int64   `json:"turn_count"`
	TokenInput            int64   `json:"token_input"`
	TokenOutput           int64   `json:"token_output"`
	TokenCacheRead        int64   `json:"token_cache_read"`
	TokenCacheWrite       int64   `json:"token_cache_write"`
	CostEstimateUsd       float64 `json:"cost_estimate_usd,omitempty"`
	ErrorCount            int64   `json:"error_count"`
}

func runExportSessions(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	opts := ports.ListSessionsOptions{
		Limit: exportLimit,
	}

	if exportExperiment != "" {
		exp, err := getExperimentByName(ctx, app.ExperimentRepo, exportExperiment)
		if err != nil {
			return err
		}
		opts.ExperimentID = &exp.ID
	}

	sessions, err := app.SessionRepo.List(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	exportData := make([]ExportSession, 0, len(sessions))
	for _, s := range sessions {
		es := ExportSession{
			ID:             s.ID,
			ProjectID:      s.ProjectID,
			Cwd:            s.Cwd,
			PermissionMode: s.PermissionMode,
			ExitReason:     s.ExitReason,
			CreatedAt:      s.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if s.ExperimentID != nil {
			es.ExperimentID = *s.ExperimentID
		}
		if s.StartedAt != nil {
			es.StartedAt = s.StartedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		if s.EndedAt != nil {
			es.EndedAt = s.EndedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		if s.DurationSeconds != nil {
			es.DurationSeconds = *s.DurationSeconds
		}

		m, err := app.MetricsRepo.GetBySessionID(ctx, s.ID)
		if err == nil && m != nil {
			es.MessageCountUser = m.MessageCountUser
			es.MessageCountAssistant = m.MessageCountAssistant
			es.TurnCount = m.TurnCount
			es.TokenInput = m.TokenInput
			es.TokenOutput = m.TokenOutput
			es.TokenCacheRead = m.TokenCacheRead
			es.TokenCacheWrite = m.TokenCacheWrite
			es.ErrorCount = m.ErrorCount
			if m.CostEstimateUSD != nil {
				es.CostEstimateUsd = *m.CostEstimateUSD
			}
		}

		exportData = append(exportData, es)
	}

	var output *os.File
	if exportOutput != "" {
		output, err = os.Create(exportOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() { _ = output.Close() }()
	} else {
		output = os.Stdout
	}

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
