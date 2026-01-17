package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/claude-watcher/internal/adapters/turso"
	sqlc "github.com/emiliopalmerini/claude-watcher/sqlc/generated"
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
  claude-watcher experiment create "minimal-prompts" --description "Testing shorter prompts" --hypothesis "Reduces token usage"`,
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

// Flags
var (
	expDescription string
	expHypothesis  string
)

func init() {
	rootCmd.AddCommand(experimentCmd)

	experimentCmd.AddCommand(experimentCreateCmd)
	experimentCmd.AddCommand(experimentListCmd)
	experimentCmd.AddCommand(experimentActivateCmd)
	experimentCmd.AddCommand(experimentDeactivateCmd)
	experimentCmd.AddCommand(experimentEndCmd)
	experimentCmd.AddCommand(experimentDeleteCmd)

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

	queries := sqlc.New(db)

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
		Description: toNullString(expDescription),
		Hypothesis:  toNullString(expHypothesis),
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

	queries := sqlc.New(db)

	experiments, err := queries.ListExperiments(ctx)
	if err != nil {
		return fmt.Errorf("failed to list experiments: %w", err)
	}

	if len(experiments) == 0 {
		fmt.Println("No experiments found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATUS\tSTARTED\tENDED\tDESCRIPTION")
	fmt.Fprintln(w, "----\t------\t-------\t-----\t-----------")

	for _, exp := range experiments {
		status := "inactive"
		if exp.IsActive == 1 {
			status = "ACTIVE"
		} else if exp.EndedAt.Valid {
			status = "ended"
		}

		started := formatDate(exp.StartedAt)
		ended := "-"
		if exp.EndedAt.Valid {
			ended = formatDate(exp.EndedAt.String)
		}

		desc := "-"
		if exp.Description.Valid && exp.Description.String != "" {
			desc = truncate(exp.Description.String, 40)
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", exp.Name, status, started, ended, desc)
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

	queries := sqlc.New(db)

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

	queries := sqlc.New(db)

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

	queries := sqlc.New(db)

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
		EndedAt:     toNullString(now),
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

	queries := sqlc.New(db)

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

func formatDate(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return s
	}
	return t.Format("2006-01-02")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
