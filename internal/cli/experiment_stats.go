package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/util"
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
	// Normalized behavior metrics
	tokensPerTurn    float64
	outputRatio      float64
	cacheHitRate     float64
	errorRate        float64
	toolCallsPerTurn float64
}

func runExperimentStats(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	name := args[0]

	exp, err := getExperimentByName(ctx, app.ExperimentRepo, name)
	if err != nil {
		return err
	}

	stats, err := app.StatsRepo.GetAggregateByExperiment(ctx, exp.ID, "1970-01-01T00:00:00Z")
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Println()
	fmt.Printf("  Experiment: %s\n", exp.Name)
	fmt.Printf("  ==============%s\n", repeatChar('=', len(exp.Name)))
	fmt.Println()

	if exp.Description != nil && *exp.Description != "" {
		fmt.Printf("  Description:  %s\n", *exp.Description)
	}
	if exp.Hypothesis != nil && *exp.Hypothesis != "" {
		fmt.Printf("  Hypothesis:   %s\n", *exp.Hypothesis)
	}
	if exp.ModelID != nil && *exp.ModelID != "" {
		fmt.Printf("  Model:        %s\n", *exp.ModelID)
	}
	if exp.PlanType != nil && *exp.PlanType != "" {
		fmt.Printf("  Plan:         %s\n", *exp.PlanType)
	}
	if exp.Notes != nil && *exp.Notes != "" {
		fmt.Printf("  Notes:        %s\n", *exp.Notes)
	}

	vars, _ := app.ExpVariableRepo.ListByExperimentID(ctx, exp.ID)
	if len(vars) > 0 {
		fmt.Printf("  Variables:\n")
		for _, v := range vars {
			fmt.Printf("    %s = %s\n", v.Key, v.Value)
		}
	}

	status := "inactive"
	if exp.IsActive {
		status = "ACTIVE"
	} else if exp.EndedAt != nil {
		status = "ended"
	}
	fmt.Printf("  Status:       %s\n", status)
	fmt.Printf("  Started:      %s\n", exp.StartedAt.Format("2006-01-02"))
	if exp.EndedAt != nil {
		fmt.Printf("  Ended:        %s\n", exp.EndedAt.Format("2006-01-02"))
	}
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

	if stats.SessionCount > 0 {
		fmt.Printf("  Efficiency\n")
		fmt.Printf("  ----------\n")
		fmt.Printf("  Tokens/session:    %s\n", util.FormatNumber(totalTokens/stats.SessionCount))
		fmt.Printf("  Cost/session:      $%.4f\n", stats.TotalCostUsd/float64(stats.SessionCount))
		fmt.Println()
	}

	// Normalized behavior metrics
	toolCalls, _ := app.StatsRepo.GetTotalToolCallsByExperiment(ctx, exp.ID)
	normalized := stats.ComputeNormalized(toolCalls)

	fmt.Printf("  Behavior\n")
	fmt.Printf("  --------\n")
	fmt.Printf("  Tokens/turn:       %s\n", util.FormatNumber(int64(normalized.TokensPerTurn)))
	fmt.Printf("  Output ratio:      %.2f\n", normalized.OutputRatio)
	fmt.Printf("  Cache hit rate:    %.1f%%\n", normalized.CacheHitRate*100)
	fmt.Printf("  Error rate:        %.2f%%\n", normalized.ErrorRate*100)
	fmt.Printf("  Tool calls/turn:   %.1f\n", normalized.ToolCallsPerTurn)
	fmt.Println()

	return nil
}

func runExperimentCompare(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	var experiments []expData

	for _, name := range args {
		exp, err := app.ExperimentRepo.GetByName(ctx, name)
		if err != nil {
			return fmt.Errorf("failed to get experiment: %w", err)
		}
		if exp == nil {
			return fmt.Errorf("experiment %q not found", name)
		}

		stats, err := app.StatsRepo.GetAggregateByExperiment(ctx, exp.ID, "1970-01-01T00:00:00Z")
		if err != nil {
			return fmt.Errorf("failed to get stats for %q: %w", name, err)
		}

		totalTokens := stats.TotalTokenInput + stats.TotalTokenOutput
		tokensPerSes := int64(0)
		costPerSes := 0.0
		if stats.SessionCount > 0 {
			tokensPerSes = totalTokens / stats.SessionCount
			costPerSes = stats.TotalCostUsd / float64(stats.SessionCount)
		}

		toolCalls, _ := app.StatsRepo.GetTotalToolCallsByExperiment(ctx, exp.ID)
		normalized := stats.ComputeNormalized(toolCalls)

		experiments = append(experiments, expData{
			name:             name,
			sessions:         stats.SessionCount,
			turns:            stats.TotalTurns,
			userMsgs:         stats.TotalUserMessages,
			assistMsgs:       stats.TotalAssistantMessages,
			tokenInput:       stats.TotalTokenInput,
			tokenOutput:      stats.TotalTokenOutput,
			cacheRead:        stats.TotalTokenCacheRead,
			cacheWrite:       stats.TotalTokenCacheWrite,
			cost:             stats.TotalCostUsd,
			errors:           stats.TotalErrors,
			totalTokens:      totalTokens,
			tokensPerSes:     tokensPerSes,
			costPerSes:       costPerSes,
			tokensPerTurn:    normalized.TokensPerTurn,
			outputRatio:      normalized.OutputRatio,
			cacheHitRate:     normalized.CacheHitRate,
			errorRate:        normalized.ErrorRate,
			toolCallsPerTurn: normalized.ToolCallsPerTurn,
		})
	}

	fmt.Println()
	fmt.Printf("  Experiment Comparison\n")
	fmt.Printf("  =====================\n")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "  METRIC\t")
	for _, e := range experiments {
		_, _ = fmt.Fprintf(w, "%s\t", e.name)
	}
	_, _ = fmt.Fprintln(w)

	_, _ = fmt.Fprintf(w, "  ------\t")
	for range experiments {
		_, _ = fmt.Fprintf(w, "------\t")
	}
	_, _ = fmt.Fprintln(w)

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
	fmt.Fprintln(w)
	printCompareRow(w, "Tokens/turn", experiments, func(e expData) string { return util.FormatNumber(int64(e.tokensPerTurn)) })
	printCompareRow(w, "Output ratio", experiments, func(e expData) string { return fmt.Sprintf("%.2f", e.outputRatio) })
	printCompareRow(w, "Cache hit rate", experiments, func(e expData) string { return fmt.Sprintf("%.1f%%", e.cacheHitRate*100) })
	printCompareRow(w, "Error rate", experiments, func(e expData) string { return fmt.Sprintf("%.2f%%", e.errorRate*100) })
	printCompareRow(w, "Tool calls/turn", experiments, func(e expData) string { return fmt.Sprintf("%.1f", e.toolCallsPerTurn) })

	_ = w.Flush()
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
