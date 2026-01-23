package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/util"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

var experimentCmd = &cobra.Command{
	Use:   "experiment",
	Short: "Manage experiments",
	Long:  `Create, list, activate, and manage experiments for A/B testing Claude usage styles.`,
}

var experimentCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new experiment",
	Long: `Create a new experiment and automatically activate it.

Examples:
  mclaude experiment create "minimal-prompts" --description "Testing shorter prompts" --hypothesis "Reduces token usage"`,
	Args: cobra.ExactArgs(1),
	RunE: runExperimentCreate,
}

var experimentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all experiments",
	RunE:  runExperimentList,
}

var experimentActivateCmd = &cobra.Command{
	Use:   "activate <name>",
	Short: "Activate an experiment",
	Long:  `Activate an experiment. Only one experiment can be active at a time.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runExperimentActivate,
}

var experimentDeactivateCmd = &cobra.Command{
	Use:   "deactivate [name]",
	Short: "Deactivate an experiment",
	Long:  `Deactivate an experiment. If no name is provided, deactivates the currently active experiment.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runExperimentDeactivate,
}

var experimentEndCmd = &cobra.Command{
	Use:   "end <name>",
	Short: "End an experiment",
	Long:  `End an experiment by setting its end date and deactivating it.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runExperimentEnd,
}

var experimentDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an experiment",
	Long:  `Delete an experiment. Sessions linked to this experiment will have their experiment_id set to NULL.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runExperimentDelete,
}

var experimentStatsCmd = &cobra.Command{
	Use:   "stats <name>",
	Short: "Show statistics for an experiment",
	Long: `Show detailed statistics for a specific experiment.

Examples:
  mclaude experiment stats "baseline"`,
	Args: cobra.ExactArgs(1),
	RunE: runExperimentStats,
}

var experimentCompareCmd = &cobra.Command{
	Use:   "compare <exp1> <exp2> [exp3...]",
	Short: "Compare statistics between experiments",
	Long: `Compare statistics side-by-side between two or more experiments.

Examples:
  mclaude experiment compare "baseline" "minimal-prompts"
  mclaude experiment compare "exp1" "exp2" "exp3"`,
	Args: cobra.MinimumNArgs(2),
	RunE: runExperimentCompare,
}

// Flags
var (
	expDescription string
	expHypothesis  string
)

// expData holds aggregated stats for an experiment (used in compare)
type expData struct {
	name         string
	sessions     int64
	turns        int64
	userMsgs     int64
	assistMsgs   int64
	tokenInput   int64
	tokenOutput  int64
	cacheRead    int64
	cacheWrite   int64
	cost         float64
	errors       int64
	totalTokens  int64
	tokensPerSes int64
	costPerSes   float64
}

func init() {
	rootCmd.AddCommand(experimentCmd)

	experimentCmd.AddCommand(experimentCreateCmd)
	experimentCmd.AddCommand(experimentListCmd)
	experimentCmd.AddCommand(experimentActivateCmd)
	experimentCmd.AddCommand(experimentDeactivateCmd)
	experimentCmd.AddCommand(experimentEndCmd)
	experimentCmd.AddCommand(experimentDeleteCmd)
	experimentCmd.AddCommand(experimentStatsCmd)
	experimentCmd.AddCommand(experimentCompareCmd)

	// Flags for create command
	experimentCreateCmd.Flags().StringVarP(&expDescription, "description", "d", "", "Description of the experiment")
	experimentCreateCmd.Flags().StringVarP(&expHypothesis, "hypothesis", "H", "", "Hypothesis to test")
}

func runExperimentCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db.DB)

	// Check if experiment with this name already exists
	existing, err := queries.GetExperimentByName(ctx, name)
	if err == nil && existing.ID != "" {
		return fmt.Errorf("experiment with name %q already exists", name)
	}

	// Deactivate all other experiments
	if err := queries.DeactivateAllExperiments(ctx); err != nil {
		return fmt.Errorf("failed to deactivate experiments: %w", err)
	}

	// Create new experiment
	now := time.Now().UTC().Format(time.RFC3339)
	exp := sqlc.CreateExperimentParams{
		ID:          uuid.New().String(),
		Name:        name,
		Description: util.NullString(expDescription),
		Hypothesis:  util.NullString(expHypothesis),
		StartedAt:   now,
		IsActive:    1,
		CreatedAt:   now,
	}

	if err := queries.CreateExperiment(ctx, exp); err != nil {
		return fmt.Errorf("failed to create experiment: %w", err)
	}

	fmt.Printf("Created and activated experiment: %s\n", name)
	return nil
}

func runExperimentList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db.DB)

	// Get experiments with stats
	expStats, err := queries.GetStatsForAllExperiments(ctx)
	if err != nil {
		return fmt.Errorf("failed to list experiments: %w", err)
	}

	if len(expStats) == 0 {
		fmt.Println("No experiments found")
		return nil
	}

	// Build a map of experiment stats for quick lookup
	statsMap := make(map[string]sqlc.GetStatsForAllExperimentsRow)
	for _, es := range expStats {
		statsMap[es.ExperimentID] = es
	}

	// Get full experiment details
	experiments, err := queries.ListExperiments(ctx)
	if err != nil {
		return fmt.Errorf("failed to list experiments: %w", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tSESSIONS\tTOKENS\tCOST\tSTARTED\tENDED")
	fmt.Fprintln(w, "----\t------\t--------\t------\t----\t-------\t-----")

	for _, exp := range experiments {
		status := "inactive"
		if exp.IsActive == 1 {
			status = "ACTIVE"
		} else if exp.EndedAt.Valid {
			status = "ended"
		}

		started := util.FormatDateISO(exp.StartedAt)
		ended := "-"
		if exp.EndedAt.Valid {
			ended = util.FormatDateISO(exp.EndedAt.String)
		}

		// Get stats for this experiment
		sessions := int64(0)
		tokens := int64(0)
		cost := 0.0
		if es, ok := statsMap[exp.ID]; ok {
			sessions = es.SessionCount
			tokens = util.ToInt64(es.TotalTokenInput) + util.ToInt64(es.TotalTokenOutput)
			cost = util.ToFloat64(es.TotalCostUsd)
		}

		fmt.Fprintf(w, "%s\t%s\t%d\t%s\t$%.2f\t%s\t%s\n",
			exp.Name, status, sessions, util.FormatNumber(tokens), cost, started, ended)
	}

	w.Flush()
	return nil
}

func runExperimentActivate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db.DB)

	// Find experiment by name
	exp, err := queries.GetExperimentByName(ctx, name)
	if err != nil {
		return fmt.Errorf("experiment %q not found", name)
	}

	if exp.IsActive == 1 {
		fmt.Printf("Experiment %q is already active\n", name)
		return nil
	}

	// Deactivate all experiments first
	if err := queries.DeactivateAllExperiments(ctx); err != nil {
		return fmt.Errorf("failed to deactivate experiments: %w", err)
	}

	// Activate the selected experiment
	if err := queries.ActivateExperiment(ctx, exp.ID); err != nil {
		return fmt.Errorf("failed to activate experiment: %w", err)
	}

	fmt.Printf("Activated experiment: %s\n", name)
	return nil
}

func runExperimentDeactivate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db.DB)

	if len(args) == 0 {
		// Deactivate the currently active experiment
		active, err := queries.GetActiveExperiment(ctx)
		if err != nil {
			fmt.Println("No active experiment to deactivate")
			return nil
		}

		if err := queries.DeactivateExperiment(ctx, active.ID); err != nil {
			return fmt.Errorf("failed to deactivate experiment: %w", err)
		}

		fmt.Printf("Deactivated experiment: %s\n", active.Name)
		return nil
	}

	// Deactivate by name
	name := args[0]
	exp, err := queries.GetExperimentByName(ctx, name)
	if err != nil {
		return fmt.Errorf("experiment %q not found", name)
	}

	if exp.IsActive == 0 {
		fmt.Printf("Experiment %q is already inactive\n", name)
		return nil
	}

	if err := queries.DeactivateExperiment(ctx, exp.ID); err != nil {
		return fmt.Errorf("failed to deactivate experiment: %w", err)
	}

	fmt.Printf("Deactivated experiment: %s\n", name)
	return nil
}

func runExperimentEnd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db.DB)

	// Find experiment by name
	exp, err := queries.GetExperimentByName(ctx, name)
	if err != nil {
		return fmt.Errorf("experiment %q not found", name)
	}

	if exp.EndedAt.Valid {
		return fmt.Errorf("experiment %q has already ended", name)
	}

	// Update experiment with end date and deactivate
	now := time.Now().UTC().Format(time.RFC3339)
	if err := queries.UpdateExperiment(ctx, sqlc.UpdateExperimentParams{
		ID:          exp.ID,
		Name:        exp.Name,
		Description: exp.Description,
		Hypothesis:  exp.Hypothesis,
		StartedAt:   exp.StartedAt,
		EndedAt:     util.NullString(now),
		IsActive:    0,
	}); err != nil {
		return fmt.Errorf("failed to end experiment: %w", err)
	}

	fmt.Printf("Ended experiment: %s\n", name)
	return nil
}

func runExperimentDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db.DB)

	// Find experiment by name
	exp, err := queries.GetExperimentByName(ctx, name)
	if err != nil {
		return fmt.Errorf("experiment %q not found", name)
	}

	// Delete the experiment (sessions will have experiment_id set to NULL via ON DELETE SET NULL)
	if err := queries.DeleteExperiment(ctx, exp.ID); err != nil {
		return fmt.Errorf("failed to delete experiment: %w", err)
	}

	fmt.Printf("Deleted experiment: %s\n", name)
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func runExperimentStats(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db.DB)

	// Get experiment by name
	exp, err := queries.GetExperimentByName(ctx, name)
	if err != nil {
		return fmt.Errorf("experiment %q not found", name)
	}

	// Get stats for this experiment
	row, err := queries.GetAggregateStatsByExperiment(ctx, sqlc.GetAggregateStatsByExperimentParams{
		ExperimentID: util.NullString(exp.ID),
		CreatedAt:    "1970-01-01T00:00:00Z", // All time
	})
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	// Print experiment details
	fmt.Println()
	fmt.Printf("  Experiment: %s\n", exp.Name)
	fmt.Printf("  ==============%s\n", repeatChar('=', len(exp.Name)))
	fmt.Println()

	if exp.Description.Valid && exp.Description.String != "" {
		fmt.Printf("  Description:  %s\n", exp.Description.String)
	}
	if exp.Hypothesis.Valid && exp.Hypothesis.String != "" {
		fmt.Printf("  Hypothesis:   %s\n", exp.Hypothesis.String)
	}

	status := "inactive"
	if exp.IsActive == 1 {
		status = "ACTIVE"
	} else if exp.EndedAt.Valid {
		status = "ended"
	}
	fmt.Printf("  Status:       %s\n", status)
	fmt.Printf("  Started:      %s\n", util.FormatDateISO(exp.StartedAt))
	if exp.EndedAt.Valid {
		fmt.Printf("  Ended:        %s\n", util.FormatDateISO(exp.EndedAt.String))
	}
	fmt.Println()

	// Print stats
	fmt.Printf("  Sessions\n")
	fmt.Printf("  --------\n")
	fmt.Printf("  Total:             %d\n", row.SessionCount)
	fmt.Printf("  Turns:             %s\n", util.FormatNumber(util.ToInt64(row.TotalTurns)))
	fmt.Printf("  User messages:     %s\n", util.FormatNumber(util.ToInt64(row.TotalUserMessages)))
	fmt.Printf("  Assistant msgs:    %s\n", util.FormatNumber(util.ToInt64(row.TotalAssistantMessages)))
	fmt.Printf("  Errors:            %d\n", util.ToInt64(row.TotalErrors))
	fmt.Println()

	fmt.Printf("  Tokens\n")
	fmt.Printf("  ------\n")
	fmt.Printf("  Input:             %s\n", util.FormatNumber(util.ToInt64(row.TotalTokenInput)))
	fmt.Printf("  Output:            %s\n", util.FormatNumber(util.ToInt64(row.TotalTokenOutput)))
	fmt.Printf("  Cache read:        %s\n", util.FormatNumber(util.ToInt64(row.TotalTokenCacheRead)))
	fmt.Printf("  Cache write:       %s\n", util.FormatNumber(util.ToInt64(row.TotalTokenCacheWrite)))
	totalTokens := util.ToInt64(row.TotalTokenInput) + util.ToInt64(row.TotalTokenOutput)
	fmt.Printf("  Total:             %s\n", util.FormatNumber(totalTokens))
	fmt.Println()

	fmt.Printf("  Cost\n")
	fmt.Printf("  ----\n")
	fmt.Printf("  Estimated:         $%.4f\n", util.ToFloat64(row.TotalCostUsd))
	fmt.Println()

	// Efficiency metrics
	if row.SessionCount > 0 {
		fmt.Printf("  Efficiency\n")
		fmt.Printf("  ----------\n")
		fmt.Printf("  Tokens/session:    %s\n", util.FormatNumber(totalTokens/row.SessionCount))
		fmt.Printf("  Cost/session:      $%.4f\n", util.ToFloat64(row.TotalCostUsd)/float64(row.SessionCount))
		fmt.Println()
	}

	return nil
}

func runExperimentCompare(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db.DB)

	// Collect stats for each experiment
	var experiments []expData

	for _, name := range args {
		exp, err := queries.GetExperimentByName(ctx, name)
		if err != nil {
			return fmt.Errorf("experiment %q not found", name)
		}

		row, err := queries.GetAggregateStatsByExperiment(ctx, sqlc.GetAggregateStatsByExperimentParams{
			ExperimentID: util.NullString(exp.ID),
			CreatedAt:    "1970-01-01T00:00:00Z",
		})
		if err != nil {
			return fmt.Errorf("failed to get stats for %q: %w", name, err)
		}

		tokenInput := util.ToInt64(row.TotalTokenInput)
		tokenOutput := util.ToInt64(row.TotalTokenOutput)
		totalTokens := tokenInput + tokenOutput
		cost := util.ToFloat64(row.TotalCostUsd)

		tokensPerSes := int64(0)
		costPerSes := 0.0
		if row.SessionCount > 0 {
			tokensPerSes = totalTokens / row.SessionCount
			costPerSes = cost / float64(row.SessionCount)
		}

		experiments = append(experiments, expData{
			name:         name,
			sessions:     row.SessionCount,
			turns:        util.ToInt64(row.TotalTurns),
			userMsgs:     util.ToInt64(row.TotalUserMessages),
			assistMsgs:   util.ToInt64(row.TotalAssistantMessages),
			tokenInput:   tokenInput,
			tokenOutput:  tokenOutput,
			cacheRead:    util.ToInt64(row.TotalTokenCacheRead),
			cacheWrite:   util.ToInt64(row.TotalTokenCacheWrite),
			cost:         cost,
			errors:       util.ToInt64(row.TotalErrors),
			totalTokens:  totalTokens,
			tokensPerSes: tokensPerSes,
			costPerSes:   costPerSes,
		})
	}

	// Print comparison table
	fmt.Println()
	fmt.Printf("  Experiment Comparison\n")
	fmt.Printf("  =====================\n")
	fmt.Println()

	// Calculate column widths
	maxNameLen := 18 // "Metric" column
	for _, e := range experiments {
		if len(e.name) > maxNameLen {
			maxNameLen = len(e.name)
		}
	}
	colWidth := maxNameLen + 2
	if colWidth < 14 {
		colWidth = 14
	}

	// Print header
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "  METRIC\t")
	for _, e := range experiments {
		fmt.Fprintf(w, "%s\t", e.name)
	}
	fmt.Fprintln(w)

	fmt.Fprintf(w, "  ------\t")
	for range experiments {
		fmt.Fprintf(w, "------\t")
	}
	fmt.Fprintln(w)

	// Print rows
	printCompareRow(w, "Sessions", experiments, func(e expData) string { return fmt.Sprintf("%d", e.sessions) })
	printCompareRow(w, "Turns", experiments, func(e expData) string { return util.FormatNumber(e.turns) })
	printCompareRow(w, "User messages", experiments, func(e expData) string { return util.FormatNumber(e.userMsgs) })
	printCompareRow(w, "Assistant msgs", experiments, func(e expData) string { return util.FormatNumber(e.assistMsgs) })
	printCompareRow(w, "Errors", experiments, func(e expData) string { return fmt.Sprintf("%d", e.errors) })
	fmt.Fprintln(w)
	printCompareRow(w, "Token input", experiments, func(e expData) string { return util.FormatNumber(e.tokenInput) })
	printCompareRow(w, "Token output", experiments, func(e expData) string { return util.FormatNumber(e.tokenOutput) })
	printCompareRow(w, "Cache read", experiments, func(e expData) string { return util.FormatNumber(e.cacheRead) })
	printCompareRow(w, "Cache write", experiments, func(e expData) string { return util.FormatNumber(e.cacheWrite) })
	printCompareRow(w, "Total tokens", experiments, func(e expData) string { return util.FormatNumber(e.totalTokens) })
	fmt.Fprintln(w)
	printCompareRow(w, "Cost", experiments, func(e expData) string { return fmt.Sprintf("$%.2f", e.cost) })
	printCompareRow(w, "Tokens/session", experiments, func(e expData) string { return util.FormatNumber(e.tokensPerSes) })
	printCompareRow(w, "Cost/session", experiments, func(e expData) string { return fmt.Sprintf("$%.4f", e.costPerSes) })

	w.Flush()
	fmt.Println()

	return nil
}

func printCompareRow(w *tabwriter.Writer, label string, experiments []expData, getValue func(expData) string) {
	fmt.Fprintf(w, "  %s\t", label)
	for _, e := range experiments {
		fmt.Fprintf(w, "%s\t", getValue(e))
	}
	fmt.Fprintln(w)
}

func repeatChar(c rune, n int) string {
	result := make([]rune, n)
	for i := range result {
		result[i] = c
	}
	return string(result)
}
