package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/domain"
)

var limitsCmd = &cobra.Command{
	Use:   "limits",
	Short: "Manage usage limits",
	Long:  `Configure and check usage limits with 5-hour rolling window support.`,
}

var limitsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List current usage and limits",
	RunE:  runLimitsList,
}

var limitsPlanCmd = &cobra.Command{
	Use:   "plan <type>",
	Short: "Set your Claude plan type",
	Long: `Set your Claude plan type for limit tracking.

Types:
  pro      - Claude Pro ($20/month, ~45 messages per 5 hours)
  max_5x   - Claude Max 5x ($100/month, ~225 messages per 5 hours)
  max_20x  - Claude Max 20x ($200/month, ~900 messages per 5 hours)

Examples:
  mclaude limits plan max_5x`,
	Args: cobra.ExactArgs(1),
	RunE: runLimitsPlan,
}

var limitsLearnCmd = &cobra.Command{
	Use:   "learn",
	Short: "Record current usage as the 100% limit",
	Long: `Record current token usage as the learned limit.

Run this when Claude Code shows you've hit your limit.
The current usage will be saved as your actual limit.`,
	RunE: runLimitsLearn,
}

var limitsCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check current usage against limits",
	Long: `Check if current usage exceeds configured limits.

Exit codes:
  0 - All limits OK
  1 - Limit exceeded
  2 - Warning threshold reached (with --warn flag)`,
	RunE: runLimitsCheck,
}

// Legacy commands (for manual limits)
var limitsSetCmd = &cobra.Command{
	Use:    "set <type> <value>",
	Short:  "Set a manual usage limit (legacy)",
	Hidden: true,
	Args:   cobra.ExactArgs(2),
	RunE:   runLimitsSet,
}

var limitsDeleteCmd = &cobra.Command{
	Use:    "delete <type>",
	Short:  "Delete a usage limit (legacy)",
	Hidden: true,
	Args:   cobra.ExactArgs(1),
	RunE:   runLimitsDelete,
}

var (
	limitsWarnThreshold float64
	limitsCheckWarn     bool
)

func init() {
	rootCmd.AddCommand(limitsCmd)

	limitsCmd.AddCommand(limitsListCmd)
	limitsCmd.AddCommand(limitsPlanCmd)
	limitsCmd.AddCommand(limitsLearnCmd)
	limitsCmd.AddCommand(limitsCheckCmd)
	limitsCmd.AddCommand(limitsSetCmd)
	limitsCmd.AddCommand(limitsDeleteCmd)

	limitsSetCmd.Flags().Float64Var(&limitsWarnThreshold, "warn", 0.8, "Warning threshold (0.0-1.0)")
	limitsCheckCmd.Flags().BoolVar(&limitsCheckWarn, "warn", false, "Exit with code 2 if warning threshold reached")
}

func runLimitsPlan(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	planType := strings.ToLower(args[0])

	if _, ok := domain.PlanPresets[planType]; !ok {
		return fmt.Errorf("invalid plan type: %s (valid: pro, max_5x, max_20x)", planType)
	}

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	planRepo := turso.NewPlanConfigRepository(db)

	config := &domain.PlanConfig{
		PlanType:    planType,
		WindowHours: 5,
	}

	if err := planRepo.Upsert(ctx, config); err != nil {
		return fmt.Errorf("failed to set plan: %w", err)
	}

	preset := domain.PlanPresets[planType]
	fmt.Printf("Plan set to %s\n", preset.Name)
	fmt.Printf("  Window: %d hours (rolling)\n", config.WindowHours)
	fmt.Printf("  Estimated limit: ~%d messages (~%s tokens)\n",
		preset.MessagesPerWindow, formatTokens(preset.TokenEstimate))
	fmt.Println("\nRun 'mclaude limits learn' when you hit your limit to record the actual token limit.")
	return nil
}

func runLimitsLearn(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	planRepo := turso.NewPlanConfigRepository(db)
	metricsRepo := turso.NewUsageMetricsRepository(db)

	config, err := planRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get plan config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("no plan configured. Run 'mclaude limits plan <type>' first")
	}

	summary, err := metricsRepo.GetRollingWindowSummary(ctx, config.WindowHours)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	if summary.TotalTokens == 0 {
		return fmt.Errorf("no token usage recorded. Make sure OTEL is configured and mclaude-otel is running")
	}

	if err := planRepo.UpdateLearnedLimit(ctx, summary.TotalTokens); err != nil {
		return fmt.Errorf("failed to save learned limit: %w", err)
	}

	fmt.Printf("Learned limit recorded: %s tokens\n", formatTokens(summary.TotalTokens))
	fmt.Println("This will be used for future limit checks.")
	return nil
}

func runLimitsList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	planRepo := turso.NewPlanConfigRepository(db)
	metricsRepo := turso.NewUsageMetricsRepository(db)

	config, err := planRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get plan config: %w", err)
	}

	if config == nil {
		fmt.Println("No plan configured")
		fmt.Println("\nUse 'mclaude limits plan <type>' to set your plan:")
		fmt.Println("  mclaude limits plan pro      # Pro ($20/month)")
		fmt.Println("  mclaude limits plan max_5x   # Max 5x ($100/month)")
		fmt.Println("  mclaude limits plan max_20x  # Max 20x ($200/month)")
		return nil
	}

	summary, err := metricsRepo.GetRollingWindowSummary(ctx, config.WindowHours)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	preset, hasPreset := domain.PlanPresets[config.PlanType]

	// Determine limit to use
	var limit float64
	var limitSource string
	if config.LearnedTokenLimit != nil {
		limit = *config.LearnedTokenLimit
		limitSource = "Learned"
	} else if hasPreset {
		limit = preset.TokenEstimate
		limitSource = "Estimated"
	}

	// Display
	planName := config.PlanType
	if hasPreset {
		planName = preset.Name
	}

	fmt.Printf("Plan: %s (%d-hour rolling window)\n\n", planName, config.WindowHours)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "METRIC\tCURRENT\tLIMIT\tUSAGE\tSTATUS")
	fmt.Fprintln(w, "------\t-------\t-----\t-----\t------")

	var percentage float64
	if limit > 0 {
		percentage = summary.TotalTokens / limit
	}
	status := getStatus(percentage, 0.8)

	limitStr := "N/A"
	usageStr := "N/A"
	if limit > 0 {
		limitStr = fmt.Sprintf("~%s", formatTokens(limit))
		usageStr = fmt.Sprintf("%.0f%%", percentage*100)
	}

	fmt.Fprintf(w, "Tokens\t%s\t%s\t%s\t%s\n",
		formatTokens(summary.TotalTokens), limitStr, usageStr, status)
	fmt.Fprintf(w, "Cost\t$%.2f\t-\t-\t-\n", summary.TotalCost)

	w.Flush()

	if limit > 0 {
		fmt.Printf("\nLimit source: %s", limitSource)
		if config.LearnedAt != nil {
			fmt.Printf(" (learned %s)", config.LearnedAt.Format("2006-01-02 15:04"))
		}
		fmt.Println()
	} else {
		fmt.Println("\nNo limit set. Run 'mclaude limits learn' when you hit your limit.")
	}

	return nil
}

func runLimitsCheck(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	planRepo := turso.NewPlanConfigRepository(db)
	metricsRepo := turso.NewUsageMetricsRepository(db)

	config, err := planRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get plan config: %w", err)
	}

	if config == nil {
		fmt.Println("No plan configured")
		return nil
	}

	summary, err := metricsRepo.GetRollingWindowSummary(ctx, config.WindowHours)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	// Determine limit
	var limit float64
	if config.LearnedTokenLimit != nil {
		limit = *config.LearnedTokenLimit
	} else if preset, ok := domain.PlanPresets[config.PlanType]; ok {
		limit = preset.TokenEstimate
	}

	if limit == 0 {
		fmt.Printf("Tokens: %s (no limit configured)\n", formatTokens(summary.TotalTokens))
		return nil
	}

	percentage := summary.TotalTokens / limit
	status := getStatus(percentage, 0.8)

	fmt.Printf("Tokens: %s / %s (%.0f%%) - %s\n",
		formatTokens(summary.TotalTokens),
		formatTokens(limit),
		percentage*100,
		status)

	if percentage >= 1.0 {
		fmt.Println("\nLimit exceeded!")
		os.Exit(1)
	}

	if percentage >= 0.8 && limitsCheckWarn {
		fmt.Println("\nWarning threshold reached")
		os.Exit(2)
	}

	fmt.Println("\nOK")
	return nil
}

// Legacy manual limit commands

func runLimitsSet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	limitType := args[0]
	valueStr := args[1]

	if !isValidLimitType(limitType) {
		return fmt.Errorf("invalid limit type: %s (valid: daily_tokens, weekly_tokens, daily_cost, weekly_cost)", limitType)
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return fmt.Errorf("invalid value: %w", err)
	}

	if value <= 0 {
		return fmt.Errorf("limit value must be positive")
	}

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	limitsRepo := turso.NewUsageLimitsRepository(db)

	limit := &domain.UsageLimit{
		ID:            limitType,
		LimitValue:    value,
		WarnThreshold: limitsWarnThreshold,
		Enabled:       true,
	}

	if err := limitsRepo.Upsert(ctx, limit); err != nil {
		return fmt.Errorf("failed to set limit: %w", err)
	}

	fmt.Printf("Set %s limit to %s\n", limitType, formatTokens(value))
	return nil
}

func runLimitsDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	limitType := args[0]

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	limitsRepo := turso.NewUsageLimitsRepository(db)

	if err := limitsRepo.Delete(ctx, limitType); err != nil {
		return fmt.Errorf("failed to delete limit: %w", err)
	}

	fmt.Printf("Deleted limit: %s\n", limitType)
	return nil
}

func isValidLimitType(t string) bool {
	return t == domain.LimitDailyTokens ||
		t == domain.LimitWeeklyTokens ||
		t == domain.LimitDailyCost ||
		t == domain.LimitWeeklyCost
}

func getStatus(percentage, warnThreshold float64) string {
	if percentage >= 1.0 {
		return "EXCEEDED"
	}
	if percentage >= warnThreshold {
		return "WARNING"
	}
	return "OK"
}

func formatTokens(tokens float64) string {
	if tokens >= 1_000_000 {
		return fmt.Sprintf("%.1fM", tokens/1_000_000)
	}
	if tokens >= 1_000 {
		return fmt.Sprintf("%.0fK", tokens/1_000)
	}
	return fmt.Sprintf("%.0f", tokens)
}
