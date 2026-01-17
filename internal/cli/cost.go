package cli

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/claude-watcher/internal/adapters/turso"
	sqlc "github.com/emiliopalmerini/claude-watcher/sqlc/generated"
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
  claude-watcher cost set claude-sonnet-4-20250514 --input 3.00 --output 15.00
  claude-watcher cost set claude-opus-4-20250514 --input 15.00 --output 75.00 --cache-read 1.50 --cache-write 18.75`,
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
	costInput      float64
	costOutput     float64
	costCacheRead  float64
	costCacheWrite float64
	costName       string
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
	costSetCmd.Flags().StringVar(&costName, "name", "", "Display name (defaults to model ID)")
	costSetCmd.MarkFlagRequired("input")
	costSetCmd.MarkFlagRequired("output")
}

func runCostList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db)

	pricing, err := queries.ListModelPricing(ctx)
	if err != nil {
		return fmt.Errorf("failed to list pricing: %w", err)
	}

	if len(pricing) == 0 {
		fmt.Println("No model pricing configured")
		fmt.Println("\nUse 'claude-watcher cost set' to add pricing")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "MODEL ID\tNAME\tINPUT/1M\tOUTPUT/1M\tCACHE R/1M\tCACHE W/1M\tDEFAULT")
	fmt.Fprintln(w, "--------\t----\t--------\t---------\t----------\t----------\t-------")

	for _, p := range pricing {
		cacheRead := "-"
		if p.CacheReadPerMillion.Valid {
			cacheRead = fmt.Sprintf("$%.2f", p.CacheReadPerMillion.Float64)
		}
		cacheWrite := "-"
		if p.CacheWritePerMillion.Valid {
			cacheWrite = fmt.Sprintf("$%.2f", p.CacheWritePerMillion.Float64)
		}
		isDefault := ""
		if p.IsDefault == 1 {
			isDefault = "*"
		}

		fmt.Fprintf(w, "%s\t%s\t$%.2f\t$%.2f\t%s\t%s\t%s\n",
			p.ID, p.DisplayName, p.InputPerMillion, p.OutputPerMillion,
			cacheRead, cacheWrite, isDefault)
	}

	w.Flush()
	return nil
}

func runCostSet(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	modelID := args[0]

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db)

	displayName := costName
	if displayName == "" {
		displayName = modelID
	}

	// Check if exists
	existing, err := queries.GetModelPricingByID(ctx, modelID)
	if err == nil && existing.ID != "" {
		// Update existing
		params := sqlc.UpdateModelPricingParams{
			ID:               modelID,
			DisplayName:      displayName,
			InputPerMillion:  costInput,
			OutputPerMillion: costOutput,
			IsDefault:        existing.IsDefault,
		}
		if costCacheRead > 0 {
			params.CacheReadPerMillion = toNullFloat64(&costCacheRead)
		}
		if costCacheWrite > 0 {
			params.CacheWritePerMillion = toNullFloat64(&costCacheWrite)
		}

		if err := queries.UpdateModelPricing(ctx, params); err != nil {
			return fmt.Errorf("failed to update pricing: %w", err)
		}
		fmt.Printf("Updated pricing for %s\n", modelID)
	} else {
		// Create new
		params := sqlc.CreateModelPricingParams{
			ID:               modelID,
			DisplayName:      displayName,
			InputPerMillion:  costInput,
			OutputPerMillion: costOutput,
			CreatedAt:        time.Now().UTC().Format(time.RFC3339),
		}
		if costCacheRead > 0 {
			params.CacheReadPerMillion = toNullFloat64(&costCacheRead)
		}
		if costCacheWrite > 0 {
			params.CacheWritePerMillion = toNullFloat64(&costCacheWrite)
		}

		// Check if this is the first pricing - make it default
		allPricing, _ := queries.ListModelPricing(ctx)
		if len(allPricing) == 0 {
			params.IsDefault = 1
		}

		if err := queries.CreateModelPricing(ctx, params); err != nil {
			return fmt.Errorf("failed to create pricing: %w", err)
		}
		fmt.Printf("Created pricing for %s\n", modelID)
	}

	return nil
}

func runCostDefault(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	modelID := args[0]

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db)

	// Check if exists
	existing, err := queries.GetModelPricingByID(ctx, modelID)
	if err != nil || existing.ID == "" {
		return fmt.Errorf("model %q not found", modelID)
	}

	if err := queries.SetDefaultModelPricing(ctx, modelID); err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	fmt.Printf("Set %s as default model for cost estimation\n", modelID)
	return nil
}

func runCostDelete(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	modelID := args[0]

	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	queries := sqlc.New(db)

	if err := queries.DeleteModelPricing(ctx, modelID); err != nil {
		return fmt.Errorf("failed to delete pricing: %w", err)
	}

	fmt.Printf("Deleted pricing for %s\n", modelID)
	return nil
}

func toNullFloat64(f *float64) sql.NullFloat64 {
	if f == nil || *f == 0 {
		return sql.NullFloat64{}
	}
	return sql.NullFloat64{Float64: *f, Valid: true}
}
