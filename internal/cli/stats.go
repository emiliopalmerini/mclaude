package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/util"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show usage statistics",
	Long: `Show summary statistics for Claude Code usage.

Examples:
  mclaude stats                          # All-time stats
  mclaude stats --period today           # Today's stats
  mclaude stats --period week            # This week's stats
  mclaude stats --experiment "baseline"  # Stats for an experiment
  mclaude stats --project <id>           # Stats for a project`,
	RunE: runStats,
}

// Flags
var (
	statsPeriod     string
	statsExperiment string
	statsProject    string
)

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.Flags().StringVarP(&statsPeriod, "period", "p", "all", "Time period: today, week, month, all")
	statsCmd.Flags().StringVarP(&statsExperiment, "experiment", "e", "", "Filter by experiment name")
	statsCmd.Flags().StringVar(&statsProject, "project", "", "Filter by project ID")
}

// Stats holds the aggregate statistics
type Stats struct {
	SessionCount           int64
	TotalUserMessages      int64
	TotalAssistantMessages int64
	TotalTurns             int64
	TotalTokenInput        int64
	TotalTokenOutput       int64
	TotalTokenCacheRead    int64
	TotalTokenCacheWrite   int64
	TotalCostUsd           float64
	TotalErrors            int64
}

func runStats(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db.DB)

	// Calculate start date based on period
	startDate := getStartDate(statsPeriod)

	var stats Stats
	var filterLabel string

	if statsExperiment != "" {
		// Get experiment ID by name
		exp, err := queries.GetExperimentByName(ctx, statsExperiment)
		if err != nil {
			return fmt.Errorf("experiment %q not found", statsExperiment)
		}

		row, err := queries.GetAggregateStatsByExperiment(ctx, sqlc.GetAggregateStatsByExperimentParams{
			ExperimentID: util.NullString(exp.ID),
			CreatedAt:    startDate,
		})
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}
		stats = statsFromExperimentRow(row)
		filterLabel = fmt.Sprintf("Experiment: %s", statsExperiment)
	} else if statsProject != "" {
		row, err := queries.GetAggregateStatsByProject(ctx, sqlc.GetAggregateStatsByProjectParams{
			ProjectID: statsProject,
			CreatedAt: startDate,
		})
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}
		stats = statsFromProjectRow(row)
		filterLabel = fmt.Sprintf("Project: %s", truncate(statsProject, 16))
	} else {
		row, err := queries.GetAggregateStats(ctx, startDate)
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}
		stats = statsFromRow(row)
		filterLabel = "All sessions"
	}

	// Get active experiment
	activeExp, _ := queries.GetActiveExperiment(ctx)
	activeExpName := "-"
	if activeExp.Name != "" {
		activeExpName = activeExp.Name
	}

	// Get top tools
	tools, _ := queries.GetTopToolsUsage(ctx, sqlc.GetTopToolsUsageParams{
		CreatedAt: startDate,
		Limit:     5,
	})

	// Print stats
	printStats(stats, filterLabel, statsPeriod, activeExpName, tools)

	return nil
}

func statsFromRow(row sqlc.GetAggregateStatsRow) Stats {
	return Stats{
		SessionCount:           row.SessionCount,
		TotalUserMessages:      util.ToInt64(row.TotalUserMessages),
		TotalAssistantMessages: util.ToInt64(row.TotalAssistantMessages),
		TotalTurns:             util.ToInt64(row.TotalTurns),
		TotalTokenInput:        util.ToInt64(row.TotalTokenInput),
		TotalTokenOutput:       util.ToInt64(row.TotalTokenOutput),
		TotalTokenCacheRead:    util.ToInt64(row.TotalTokenCacheRead),
		TotalTokenCacheWrite:   util.ToInt64(row.TotalTokenCacheWrite),
		TotalCostUsd:           util.ToFloat64(row.TotalCostUsd),
		TotalErrors:            util.ToInt64(row.TotalErrors),
	}
}

func statsFromExperimentRow(row sqlc.GetAggregateStatsByExperimentRow) Stats {
	return Stats{
		SessionCount:           row.SessionCount,
		TotalUserMessages:      util.ToInt64(row.TotalUserMessages),
		TotalAssistantMessages: util.ToInt64(row.TotalAssistantMessages),
		TotalTurns:             util.ToInt64(row.TotalTurns),
		TotalTokenInput:        util.ToInt64(row.TotalTokenInput),
		TotalTokenOutput:       util.ToInt64(row.TotalTokenOutput),
		TotalTokenCacheRead:    util.ToInt64(row.TotalTokenCacheRead),
		TotalTokenCacheWrite:   util.ToInt64(row.TotalTokenCacheWrite),
		TotalCostUsd:           util.ToFloat64(row.TotalCostUsd),
		TotalErrors:            util.ToInt64(row.TotalErrors),
	}
}

func statsFromProjectRow(row sqlc.GetAggregateStatsByProjectRow) Stats {
	return Stats{
		SessionCount:           row.SessionCount,
		TotalUserMessages:      util.ToInt64(row.TotalUserMessages),
		TotalAssistantMessages: util.ToInt64(row.TotalAssistantMessages),
		TotalTurns:             util.ToInt64(row.TotalTurns),
		TotalTokenInput:        util.ToInt64(row.TotalTokenInput),
		TotalTokenOutput:       util.ToInt64(row.TotalTokenOutput),
		TotalTokenCacheRead:    util.ToInt64(row.TotalTokenCacheRead),
		TotalTokenCacheWrite:   util.ToInt64(row.TotalTokenCacheWrite),
		TotalCostUsd:           util.ToFloat64(row.TotalCostUsd),
		TotalErrors:            util.ToInt64(row.TotalErrors),
	}
}

func getStartDate(period string) string {
	now := time.Now().UTC()
	var start time.Time

	switch period {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case "week":
		// Start of current week (Monday)
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, time.UTC)
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		// All time - use Unix epoch
		start = time.Unix(0, 0)
	}

	return start.Format(time.RFC3339)
}

func printStats(stats Stats, filterLabel, period, activeExp string, tools []sqlc.GetTopToolsUsageRow) {
	periodLabel := "All time"
	switch period {
	case "today":
		periodLabel = "Today"
	case "week":
		periodLabel = "This week"
	case "month":
		periodLabel = "This month"
	}

	fmt.Println()
	fmt.Printf("  mclaude Stats\n")
	fmt.Printf("  =====================\n")
	fmt.Println()

	fmt.Printf("  Period:            %s\n", periodLabel)
	fmt.Printf("  Filter:            %s\n", filterLabel)
	fmt.Printf("  Active experiment: %s\n", activeExp)
	fmt.Println()

	fmt.Printf("  Sessions\n")
	fmt.Printf("  --------\n")
	fmt.Printf("  Total:             %d\n", stats.SessionCount)
	fmt.Printf("  Turns:             %s\n", util.FormatNumber(stats.TotalTurns))
	fmt.Printf("  User messages:     %s\n", util.FormatNumber(stats.TotalUserMessages))
	fmt.Printf("  Assistant msgs:    %s\n", util.FormatNumber(stats.TotalAssistantMessages))
	fmt.Printf("  Errors:            %d\n", stats.TotalErrors)
	fmt.Println()

	fmt.Printf("  Tokens\n")
	fmt.Printf("  ------\n")
	fmt.Printf("  Input:             %s\n", util.FormatNumber(stats.TotalTokenInput))
	fmt.Printf("  Output:            %s\n", util.FormatNumber(stats.TotalTokenOutput))
	fmt.Printf("  Cache read:        %s\n", util.FormatNumber(stats.TotalTokenCacheRead))
	fmt.Printf("  Cache write:       %s\n", util.FormatNumber(stats.TotalTokenCacheWrite))
	totalTokens := stats.TotalTokenInput + stats.TotalTokenOutput
	fmt.Printf("  Total:             %s\n", util.FormatNumber(totalTokens))
	fmt.Println()

	fmt.Printf("  Cost\n")
	fmt.Printf("  ----\n")
	fmt.Printf("  Estimated:         $%.4f\n", stats.TotalCostUsd)
	fmt.Println()

	if len(tools) > 0 {
		fmt.Printf("  Top Tools\n")
		fmt.Printf("  ---------\n")
		for _, tool := range tools {
			invocations := int64(0)
			if tool.TotalInvocations.Valid {
				invocations = int64(tool.TotalInvocations.Float64)
			}
			fmt.Printf("  %-18s %s calls\n", tool.ToolName, util.FormatNumber(invocations))
		}
		fmt.Println()
	}
}
