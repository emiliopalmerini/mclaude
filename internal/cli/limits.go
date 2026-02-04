package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/adapters/prometheus"
	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/ports"
	"github.com/emiliopalmerini/mclaude/internal/util"
)

var limitsCmd = &cobra.Command{
	Use:   "limits",
	Short: "Manage usage limits",
	Long:  `Configure and check usage limits with 5-hour and weekly rolling window support.`,
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
	Use:   "learn <type>",
	Short: "Record current usage as the 100% limit for 5-hour limit or weekly limit",
	Long: `Record current token usage as the learned limit.

Run this when Claude Code shows you've hit your limit.
The current usage will be saved as your actual limit.

Use --weekly to record the weekly limit instead of the 5-hour limit.`,
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

var (
	limitsCheckWarn   bool
	limitsLearnWeekly bool
	limitsSource      string // "local", "prometheus", "auto"
)

func init() {
	rootCmd.AddCommand(limitsCmd)

	limitsCmd.AddCommand(limitsListCmd)
	limitsCmd.AddCommand(limitsPlanCmd)
	limitsCmd.AddCommand(limitsLearnCmd)
	limitsCmd.AddCommand(limitsCheckCmd)

	limitsCheckCmd.Flags().BoolVar(&limitsCheckWarn, "warn", false, "Exit with code 2 if warning threshold reached")
	limitsLearnCmd.Flags().BoolVar(&limitsLearnWeekly, "weekly", false, "Record weekly limit instead of 5-hour limit")

	// Add --source flag to list and check commands
	limitsListCmd.Flags().StringVar(&limitsSource, "source", "auto", "Data source: local, prometheus, auto (default auto)")
	limitsCheckCmd.Flags().StringVar(&limitsSource, "source", "auto", "Data source: local, prometheus, auto (default auto)")
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

	planRepo := turso.NewPlanConfigRepository(db.DB)

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
		preset.MessagesPerWindow, util.FormatTokens(preset.TokenEstimate))
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

	planRepo := turso.NewPlanConfigRepository(db.DB)

	config, err := planRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get plan config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("no plan configured. Run 'mclaude limits plan <type>' first")
	}

	if limitsLearnWeekly {
		// Learn weekly limit
		summary, err := planRepo.GetWeeklyWindowSummary(ctx)
		if err != nil {
			return fmt.Errorf("failed to get weekly usage: %w", err)
		}

		if summary.TotalTokens == 0 {
			return fmt.Errorf("no token usage recorded in the last 7 days")
		}

		if err := planRepo.UpdateWeeklyLearnedLimit(ctx, summary.TotalTokens); err != nil {
			return fmt.Errorf("failed to save weekly learned limit: %w", err)
		}

		fmt.Printf("Weekly learned limit recorded: %s tokens\n", util.FormatTokens(summary.TotalTokens))
		fmt.Println("This will be used for future weekly limit checks.")
	} else {
		// Learn 5-hour limit
		summary, err := planRepo.GetRollingWindowSummary(ctx, config.WindowHours)
		if err != nil {
			return fmt.Errorf("failed to get usage: %w", err)
		}

		if summary.TotalTokens == 0 {
			return fmt.Errorf("no token usage recorded in the last %d hours", config.WindowHours)
		}

		if err := planRepo.UpdateLearnedLimit(ctx, summary.TotalTokens); err != nil {
			return fmt.Errorf("failed to save learned limit: %w", err)
		}

		fmt.Printf("Learned limit recorded: %s tokens\n", util.FormatTokens(summary.TotalTokens))
		fmt.Println("This will be used for future limit checks.")
	}
	return nil
}

func runLimitsList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	planRepo := turso.NewPlanConfigRepository(db.DB)

	config, err := planRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get plan config: %w", err)
	}

	if config != nil {
		// Reset windows if expired before showing usage
		now := time.Now()
		if _, err := planRepo.ResetWindowIfExpired(ctx, now); err != nil {
			return fmt.Errorf("failed to reset 5-hour window: %w", err)
		}
		if _, err := planRepo.ResetWeeklyWindowIfExpired(ctx, now); err != nil {
			return fmt.Errorf("failed to reset weekly window: %w", err)
		}
	}

	if config == nil {
		fmt.Println("No plan configured")
		fmt.Println("\nUse 'mclaude limits plan <type>' to set your plan:")
		fmt.Println("  mclaude limits plan pro      # Pro ($20/month)")
		fmt.Println("  mclaude limits plan max_5x   # Max 5x ($100/month)")
		fmt.Println("  mclaude limits plan max_20x  # Max 20x ($200/month)")
		return nil
	}

	preset, hasPreset := domain.PlanPresets[config.PlanType]
	weeklyPreset, hasWeeklyPreset := domain.WeeklyPlanPresets[config.PlanType]

	// Display plan name
	planName := config.PlanType
	if hasPreset {
		planName = preset.Name
	}

	// Initialize Prometheus client for real-time data
	promClient := getPrometheusClient()

	// Try Prometheus first for real-time data if source allows
	var promUsage *ports.UsageWindow
	var promSource bool
	if limitsSource != "local" && promClient.IsAvailable(ctx) {
		usage, err := promClient.GetRollingWindowUsage(ctx, config.WindowHours)
		if err == nil && usage.Available {
			promUsage = usage
			promSource = true
		}
	}

	fmt.Printf("Plan: %s", planName)
	if promSource {
		fmt.Printf(" [Source: Prometheus]")
	} else {
		fmt.Printf(" [Source: Local DB]")
	}
	fmt.Println()
	fmt.Println()

	// === 5-Hour Window ===
	var fiveHourTokens float64
	if promSource && promUsage != nil {
		fiveHourTokens = promUsage.TotalTokens
	} else {
		summary, err := planRepo.GetRollingWindowSummary(ctx, config.WindowHours)
		if err != nil {
			return fmt.Errorf("failed to get usage: %w", err)
		}
		fiveHourTokens = summary.TotalTokens
	}

	var limit float64
	var tokenLimitSource string
	if config.LearnedTokenLimit != nil {
		limit = *config.LearnedTokenLimit
		tokenLimitSource = "Learned"
	} else if hasPreset {
		limit = preset.TokenEstimate
		tokenLimitSource = "Estimated"
	}

	var percentage float64
	if limit > 0 {
		percentage = fiveHourTokens / limit
	}
	status := domain.GetStatus(percentage, 0.8)

	fmt.Printf("5-Hour Window:\n")
	fmt.Printf("  Tokens: %s", util.FormatTokens(fiveHourTokens))
	if limit > 0 {
		fmt.Printf(" / ~%s (%.0f%%)", util.FormatTokens(limit), percentage*100)
	}
	fmt.Println()
	fmt.Printf("  Status: %s", status)
	if limit > 0 {
		fmt.Printf(" [%s]", tokenLimitSource)
	}
	fmt.Println()

	// === Weekly Window ===
	var weeklyTokens float64
	if promSource {
		// Query Prometheus for 7-day window
		weeklyUsage, err := promClient.GetRollingWindowUsage(ctx, 168) // 7 days = 168 hours
		if err == nil && weeklyUsage.Available {
			weeklyTokens = weeklyUsage.TotalTokens
		} else {
			// Fall back to local DB for weekly
			weeklySummary, err := planRepo.GetWeeklyWindowSummary(ctx)
			if err != nil {
				return fmt.Errorf("failed to get weekly usage: %w", err)
			}
			weeklyTokens = weeklySummary.TotalTokens
		}
	} else {
		weeklySummary, err := planRepo.GetWeeklyWindowSummary(ctx)
		if err != nil {
			return fmt.Errorf("failed to get weekly usage: %w", err)
		}
		weeklyTokens = weeklySummary.TotalTokens
	}

	var weeklyLimit float64
	var weeklyLimitSource string
	if config.WeeklyLearnedTokenLimit != nil {
		weeklyLimit = *config.WeeklyLearnedTokenLimit
		weeklyLimitSource = "Learned"
	} else if hasWeeklyPreset {
		weeklyLimit = weeklyPreset.TokenEstimate
		weeklyLimitSource = "Estimated"
	}

	var weeklyPercentage float64
	if weeklyLimit > 0 {
		weeklyPercentage = weeklyTokens / weeklyLimit
	}
	weeklyStatus := domain.GetStatus(weeklyPercentage, 0.8)

	fmt.Printf("\nWeekly Window (7 days):\n")
	fmt.Printf("  Tokens: %s", util.FormatTokens(weeklyTokens))
	if weeklyLimit > 0 {
		fmt.Printf(" / ~%s (%.0f%%)", util.FormatTokens(weeklyLimit), weeklyPercentage*100)
	}
	fmt.Println()
	fmt.Printf("  Status: %s", weeklyStatus)
	if weeklyLimit > 0 {
		fmt.Printf(" [%s]", weeklyLimitSource)
	}
	fmt.Println()

	// Hints
	if limit == 0 || weeklyLimit == 0 {
		fmt.Println("\nTip: Run 'mclaude limits learn' when you hit your 5-hour limit.")
		fmt.Println("     Run 'mclaude limits learn --weekly' when you hit your weekly limit.")
	}

	return nil
}

// getPrometheusClient returns a Prometheus client (real or noop).
func getPrometheusClient() ports.PrometheusClient {
	cfg := prometheus.LoadConfig()
	client, err := prometheus.NewClient(cfg)
	if err != nil {
		return prometheus.NewNoOpClient()
	}
	return client
}

func runLimitsCheck(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	planRepo := turso.NewPlanConfigRepository(db.DB)

	config, err := planRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get plan config: %w", err)
	}

	if config == nil {
		fmt.Println("No plan configured")
		return nil
	}

	// Reset windows if expired before checking usage
	now := time.Now()
	if _, err := planRepo.ResetWindowIfExpired(ctx, now); err != nil {
		return fmt.Errorf("failed to reset 5-hour window: %w", err)
	}
	if _, err := planRepo.ResetWeeklyWindowIfExpired(ctx, now); err != nil {
		return fmt.Errorf("failed to reset weekly window: %w", err)
	}

	// Check 5-hour window
	summary, err := planRepo.GetRollingWindowSummary(ctx, config.WindowHours)
	if err != nil {
		return fmt.Errorf("failed to get usage: %w", err)
	}

	var limit float64
	if config.LearnedTokenLimit != nil {
		limit = *config.LearnedTokenLimit
	} else if preset, ok := domain.PlanPresets[config.PlanType]; ok {
		limit = preset.TokenEstimate
	}

	var percentage float64
	if limit > 0 {
		percentage = summary.TotalTokens / limit
	}
	status := domain.GetStatus(percentage, 0.8)

	fmt.Printf("5-Hour: %s", util.FormatTokens(summary.TotalTokens))
	if limit > 0 {
		fmt.Printf(" / %s (%.0f%%) - %s", util.FormatTokens(limit), percentage*100, status)
	}
	fmt.Println()

	// Check weekly window
	weeklySummary, err := planRepo.GetWeeklyWindowSummary(ctx)
	if err != nil {
		return fmt.Errorf("failed to get weekly usage: %w", err)
	}

	var weeklyLimit float64
	if config.WeeklyLearnedTokenLimit != nil {
		weeklyLimit = *config.WeeklyLearnedTokenLimit
	} else if preset, ok := domain.WeeklyPlanPresets[config.PlanType]; ok {
		weeklyLimit = preset.TokenEstimate
	}

	var weeklyPercentage float64
	if weeklyLimit > 0 {
		weeklyPercentage = weeklySummary.TotalTokens / weeklyLimit
	}
	weeklyStatus := domain.GetStatus(weeklyPercentage, 0.8)

	fmt.Printf("Weekly: %s", util.FormatTokens(weeklySummary.TotalTokens))
	if weeklyLimit > 0 {
		fmt.Printf(" / %s (%.0f%%) - %s", util.FormatTokens(weeklyLimit), weeklyPercentage*100, weeklyStatus)
	}
	fmt.Println()

	// Exit codes: either limit exceeded = 1, either warning = 2 (with --warn)
	if percentage >= 1.0 || weeklyPercentage >= 1.0 {
		if percentage >= 1.0 {
			fmt.Println("\n5-hour limit exceeded!")
		}
		if weeklyPercentage >= 1.0 {
			fmt.Println("\nWeekly limit exceeded!")
		}
		os.Exit(1)
	}

	if limitsCheckWarn && (percentage >= 0.8 || weeklyPercentage >= 0.8) {
		fmt.Println("\nWarning threshold reached")
		os.Exit(2)
	}

	fmt.Println("\nOK")
	return nil
}
