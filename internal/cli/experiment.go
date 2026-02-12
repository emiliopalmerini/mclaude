package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/util"
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
	expModel       string
	expPlan        string
	expNotes       string
	expVars        []string
)

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
	experimentCreateCmd.Flags().StringVarP(&expModel, "model", "m", "", "Model ID used for this experiment (e.g. claude-opus-4-6)")
	experimentCreateCmd.Flags().StringVarP(&expPlan, "plan", "p", "", "Plan type (e.g. pro, max_5x, max_20x)")
	experimentCreateCmd.Flags().StringVarP(&expNotes, "notes", "n", "", "Free-form notes about methodology")
	experimentCreateCmd.Flags().StringArrayVar(&expVars, "var", nil, "Variable key=value pair (can be repeated)")
}

func runExperimentCreate(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	existing, err := app.ExperimentRepo.GetByName(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check experiment: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("experiment with name %q already exists", name)
	}

	if err := app.ExperimentRepo.DeactivateAll(ctx); err != nil {
		return fmt.Errorf("failed to deactivate experiments: %w", err)
	}

	now := time.Now().UTC()
	exp := &domain.Experiment{
		ID:        uuid.New().String(),
		Name:      name,
		StartedAt: now,
		IsActive:  true,
		CreatedAt: now,
	}
	if expDescription != "" {
		exp.Description = &expDescription
	}
	if expHypothesis != "" {
		exp.Hypothesis = &expHypothesis
	}
	if expModel != "" {
		exp.ModelID = &expModel
	}
	if expPlan != "" {
		exp.PlanType = &expPlan
	}
	if expNotes != "" {
		exp.Notes = &expNotes
	}

	if err := app.ExperimentRepo.Create(ctx, exp); err != nil {
		return fmt.Errorf("failed to create experiment: %w", err)
	}

	for _, v := range expVars {
		key, value, ok := strings.Cut(v, "=")
		if !ok {
			return fmt.Errorf("invalid variable format %q, expected key=value", v)
		}
		if err := app.ExpVariableRepo.Set(ctx, exp.ID, key, value); err != nil {
			return fmt.Errorf("failed to set variable %q: %w", key, err)
		}
	}

	fmt.Printf("Created and activated experiment: %s\n", name)
	return nil
}

func runExperimentList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	expStats, err := app.StatsRepo.GetAllExperimentStats(ctx)
	if err != nil {
		return fmt.Errorf("failed to get experiment stats: %w", err)
	}

	experiments, err := app.ExperimentRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list experiments: %w", err)
	}

	if len(experiments) == 0 {
		fmt.Println("No experiments found")
		return nil
	}

	statsMap := make(map[string]domain.ExperimentStats)
	for _, es := range expStats {
		statsMap[es.ExperimentID] = es
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tSTATUS\tMODEL\tPLAN\tSESSIONS\tTOKENS\tCOST\tSTARTED\tENDED")
	_, _ = fmt.Fprintln(w, "----\t------\t-----\t----\t--------\t------\t----\t-------\t-----")

	for _, exp := range experiments {
		status := "inactive"
		if exp.IsActive {
			status = "ACTIVE"
		} else if exp.EndedAt != nil {
			status = "ended"
		}

		started := exp.StartedAt.Format("2006-01-02")
		ended := "-"
		if exp.EndedAt != nil {
			ended = exp.EndedAt.Format("2006-01-02")
		}

		sessions := int64(0)
		tokens := int64(0)
		cost := 0.0
		if es, ok := statsMap[exp.ID]; ok {
			sessions = es.SessionCount
			tokens = es.TotalTokenInput + es.TotalTokenOutput
			cost = es.TotalCostUsd
		}

		model := "-"
		if exp.ModelID != nil {
			model = *exp.ModelID
		}
		plan := "-"
		if exp.PlanType != nil {
			plan = *exp.PlanType
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\t$%.2f\t%s\t%s\n",
			exp.Name, status, model, plan, sessions, util.FormatNumber(tokens), cost, started, ended)
	}

	_ = w.Flush()
	return nil
}

func runExperimentDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	exp, err := getExperimentByName(ctx, app.ExperimentRepo, name)
	if err != nil {
		return err
	}

	if err := app.ExperimentRepo.Delete(ctx, exp.ID); err != nil {
		return fmt.Errorf("failed to delete experiment: %w", err)
	}

	fmt.Printf("Deleted experiment: %s\n", name)
	return nil
}
