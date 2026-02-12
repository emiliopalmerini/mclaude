package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Manage model pricing",
	Long:  `Configure model pricing for cost estimation.`,
}

var costListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured model pricing",
	RunE:  runCostList,
}

var costSetCmd = &cobra.Command{
	Use:   "set <model-id>",
	Short: "Set model pricing",
	Long: `Set pricing for a model (USD per 1M tokens).

Examples:
  mclaude cost set claude-sonnet-4-20250514 --input 3.00 --output 15.00
  mclaude cost set claude-opus-4-6-20260115 --input 5.00 --output 25.00 --cache-read 0.50 --cache-write 6.25 --long-input 10.00 --long-output 37.50
  mclaude cost set claude-opus-4-20250514 --input 15.00 --output 75.00 --cache-read 1.50 --cache-write 18.75`,
	Args: cobra.ExactArgs(1),
	RunE: runCostSet,
}

var costDefaultCmd = &cobra.Command{
	Use:   "default <model-id>",
	Short: "Set the default model for cost estimation",
	Args:  cobra.ExactArgs(1),
	RunE:  runCostDefault,
}

var costDeleteCmd = &cobra.Command{
	Use:   "delete <model-id>",
	Short: "Delete model pricing",
	Args:  cobra.ExactArgs(1),
	RunE:  runCostDelete,
}

// Flags
var (
	costInput         float64
	costOutput        float64
	costCacheRead     float64
	costCacheWrite    float64
	costName          string
	costLongInput     float64
	costLongOutput    float64
	costLongThreshold int64
)

func init() {
	rootCmd.AddCommand(costCmd)

	costCmd.AddCommand(costListCmd)
	costCmd.AddCommand(costSetCmd)
	costCmd.AddCommand(costDefaultCmd)
	costCmd.AddCommand(costDeleteCmd)

	costSetCmd.Flags().Float64Var(&costInput, "input", 0, "Input tokens cost per 1M (required)")
	costSetCmd.Flags().Float64Var(&costOutput, "output", 0, "Output tokens cost per 1M (required)")
	costSetCmd.Flags().Float64Var(&costCacheRead, "cache-read", 0, "Cache read tokens cost per 1M")
	costSetCmd.Flags().Float64Var(&costCacheWrite, "cache-write", 0, "Cache write tokens cost per 1M")
	costSetCmd.Flags().Float64Var(&costLongInput, "long-input", 0, "Long context input cost per 1M (>200K tokens)")
	costSetCmd.Flags().Float64Var(&costLongOutput, "long-output", 0, "Long context output cost per 1M (>200K tokens)")
	costSetCmd.Flags().Int64Var(&costLongThreshold, "long-threshold", 200000, "Input token threshold for long context pricing")
	costSetCmd.Flags().StringVar(&costName, "name", "", "Display name (defaults to model ID)")
	_ = costSetCmd.MarkFlagRequired("input")
	_ = costSetCmd.MarkFlagRequired("output")
}

func runCostList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	pricing, err := app.PricingRepo.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list pricing: %w", err)
	}

	if len(pricing) == 0 {
		fmt.Println("No model pricing configured")
		fmt.Println("\nUse 'mclaude cost set' to add pricing")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "MODEL ID\tNAME\tINPUT/1M\tOUTPUT/1M\tCACHE R/1M\tCACHE W/1M\tLONG IN/1M\tLONG OUT/1M\tDEFAULT")
	_, _ = fmt.Fprintln(w, "--------\t----\t--------\t---------\t----------\t----------\t----------\t-----------\t-------")

	for _, p := range pricing {
		cacheRead := "-"
		if p.CacheReadPerMillion != nil {
			cacheRead = fmt.Sprintf("$%.2f", *p.CacheReadPerMillion)
		}
		cacheWrite := "-"
		if p.CacheWritePerMillion != nil {
			cacheWrite = fmt.Sprintf("$%.2f", *p.CacheWritePerMillion)
		}
		longInput := "-"
		if p.LongContextInputPerMillion != nil {
			longInput = fmt.Sprintf("$%.2f", *p.LongContextInputPerMillion)
		}
		longOutput := "-"
		if p.LongContextOutputPerMillion != nil {
			longOutput = fmt.Sprintf("$%.2f", *p.LongContextOutputPerMillion)
		}
		isDefault := ""
		if p.IsDefault {
			isDefault = "*"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t$%.2f\t$%.2f\t%s\t%s\t%s\t%s\t%s\n",
			p.ID, p.DisplayName, p.InputPerMillion, p.OutputPerMillion,
			cacheRead, cacheWrite, longInput, longOutput, isDefault)
	}

	_ = w.Flush()
	return nil
}

func runCostSet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	modelID := args[0]

	displayName := costName
	if displayName == "" {
		displayName = modelID
	}

	pricing := &domain.ModelPricing{
		ID:              modelID,
		DisplayName:     displayName,
		InputPerMillion: costInput,
		OutputPerMillion: costOutput,
		CreatedAt:       time.Now().UTC(),
	}

	if costCacheRead > 0 {
		pricing.CacheReadPerMillion = &costCacheRead
	}
	if costCacheWrite > 0 {
		pricing.CacheWritePerMillion = &costCacheWrite
	}
	if costLongInput > 0 {
		pricing.LongContextInputPerMillion = &costLongInput
	}
	if costLongOutput > 0 {
		pricing.LongContextOutputPerMillion = &costLongOutput
	}
	if costLongInput > 0 || costLongOutput > 0 {
		pricing.LongContextThreshold = &costLongThreshold
	}

	existing, _ := app.PricingRepo.GetByID(ctx, modelID)
	if existing != nil {
		pricing.IsDefault = existing.IsDefault
		pricing.CreatedAt = existing.CreatedAt
		if err := app.PricingRepo.Update(ctx, pricing); err != nil {
			return fmt.Errorf("failed to update pricing: %w", err)
		}
		fmt.Printf("Updated pricing for %s\n", modelID)
	} else {
		allPricing, _ := app.PricingRepo.List(ctx)
		if len(allPricing) == 0 {
			pricing.IsDefault = true
		}

		if err := app.PricingRepo.Create(ctx, pricing); err != nil {
			return fmt.Errorf("failed to create pricing: %w", err)
		}
		fmt.Printf("Created pricing for %s\n", modelID)
	}

	return nil
}

func runCostDefault(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	modelID := args[0]

	existing, _ := app.PricingRepo.GetByID(ctx, modelID)
	if existing == nil {
		return fmt.Errorf("model %q not found", modelID)
	}

	if err := app.PricingRepo.SetDefault(ctx, modelID); err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	fmt.Printf("Set %s as default model for cost estimation\n", modelID)
	return nil
}

func runCostDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	modelID := args[0]

	if err := app.PricingRepo.Delete(ctx, modelID); err != nil {
		return fmt.Errorf("failed to delete pricing: %w", err)
	}

	fmt.Printf("Deleted pricing for %s\n", modelID)
	return nil
}
