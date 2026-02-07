package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/adapters/storage"
	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/parser"
)

// testDBOverride allows tests to inject a database connection.
// When set, processRecordInput uses this instead of creating a new connection.
var testDBOverride *sql.DB

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record session data from Claude Code hook",
	Long: `Reads session data from stdin (Claude Code SessionEnd hook),
parses the transcript, and saves all data to the database.

This command is designed to be called from a Claude Code hook with
"async": true so it runs in the background:

  {
    "hooks": {
      "SessionEnd": [
        {
          "hooks": [
            {
              "type": "command",
              "command": "mclaude record",
              "async": true
            }
          ]
        }
      ]
    }
  }`,
	RunE: runRecord,
}

func runRecord(cmd *cobra.Command, args []string) error {
	// Read hook input from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	var hookInput domain.HookInput
	if err := json.Unmarshal(input, &hookInput); err != nil {
		return fmt.Errorf("failed to parse hook input: %w", err)
	}

	return processRecordInput(&hookInput)
}

func processRecordInput(hookInput *domain.HookInput) error {
	ctx := context.Background()

	// Use test database if set, otherwise connect to real database
	var sqlDB *sql.DB
	var tursoDB *turso.DB // Keep reference for Sync() call
	var closeDB func()

	if testDBOverride != nil {
		sqlDB = testDBOverride
		closeDB = func() {} // Don't close test database
	} else {
		db, err := turso.NewDB()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		sqlDB = db.DB
		tursoDB = db
		closeDB = func() { db.Close() }
	}
	defer closeDB()

	// Initialize repositories
	projectRepo := turso.NewProjectRepository(sqlDB)
	experimentRepo := turso.NewExperimentRepository(sqlDB)
	sessionRepo := turso.NewSessionRepository(sqlDB)
	metricsRepo := turso.NewSessionMetricsRepository(sqlDB)
	toolRepo := turso.NewSessionToolRepository(sqlDB)
	fileRepo := turso.NewSessionFileRepository(sqlDB)
	commandRepo := turso.NewSessionCommandRepository(sqlDB)
	subagentRepo := turso.NewSessionSubagentRepository(sqlDB)
	pricingRepo := turso.NewPricingRepository(sqlDB)
	qualityRepo := turso.NewSessionQualityRepository(sqlDB)
	planConfigRepo := turso.NewPlanConfigRepository(sqlDB)

	// Initialize transcript storage
	transcriptStorage, err := storage.NewTranscriptStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize transcript storage: %w", err)
	}

	// Get or create project
	project, err := projectRepo.GetOrCreate(ctx, hookInput.Cwd)
	if err != nil {
		return fmt.Errorf("failed to get/create project: %w", err)
	}

	// Get active experiment (if any)
	activeExperiment, err := experimentRepo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active experiment: %w", err)
	}

	// Parse transcript
	parsed, err := parser.ParseTranscript(hookInput.SessionID, hookInput.TranscriptPath)
	if err != nil {
		return fmt.Errorf("failed to parse transcript: %w", err)
	}

	// Reset usage windows if expired
	if parsed.StartedAt != nil {
		if reset, err := planConfigRepo.ResetWindowIfExpired(ctx, *parsed.StartedAt); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to check window reset: %v\n", err)
		} else if reset {
			fmt.Println("5-hour usage window reset")
		}
		if reset, err := planConfigRepo.ResetWeeklyWindowIfExpired(ctx, *parsed.StartedAt); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to check weekly window reset: %v\n", err)
		} else if reset {
			fmt.Println("Weekly usage window reset")
		}
	}

	// Store transcript copy
	storedPath, err := transcriptStorage.Store(ctx, hookInput.SessionID, hookInput.TranscriptPath)
	if err != nil {
		// Log but don't fail - transcript storage is not critical
		fmt.Fprintf(os.Stderr, "warning: failed to store transcript copy: %v\n", err)
	}

	// Set model ID from parsed transcript
	parsed.Metrics.ModelID = parsed.ModelID

	// Calculate cost estimate using model-specific pricing if available
	var costEstimate *float64
	var pricing *domain.ModelPricing

	// Try to get pricing for the specific model used
	if parsed.ModelID != nil {
		pricing, _ = pricingRepo.GetByID(ctx, *parsed.ModelID)
	}

	// Fall back to default pricing if model-specific pricing not found
	if pricing == nil {
		pricing, _ = pricingRepo.GetDefault(ctx)
	}

	if pricing != nil {
		cost := pricing.CalculateCost(
			parsed.Metrics.TokenInput,
			parsed.Metrics.TokenOutput,
			parsed.Metrics.TokenCacheRead,
			parsed.Metrics.TokenCacheWrite,
		)
		costEstimate = &cost
	}
	parsed.Metrics.CostEstimateUSD = costEstimate

	// Calculate duration
	var durationSeconds *int64
	if parsed.StartedAt != nil && parsed.EndedAt != nil {
		dur := int64(parsed.EndedAt.Sub(*parsed.StartedAt).Seconds())
		durationSeconds = &dur
	}

	// Build session
	session := &domain.Session{
		ID:              hookInput.SessionID,
		ProjectID:       project.ID,
		TranscriptPath:  hookInput.TranscriptPath,
		Cwd:             hookInput.Cwd,
		PermissionMode:  hookInput.PermissionMode,
		ExitReason:      hookInput.Reason,
		StartedAt:       parsed.StartedAt,
		EndedAt:         parsed.EndedAt,
		DurationSeconds: durationSeconds,
		CreatedAt:       time.Now().UTC(),
	}

	if storedPath != "" {
		session.TranscriptStoredPath = &storedPath
	}

	if activeExperiment != nil {
		session.ExperimentID = &activeExperiment.ID
	}

	// Save session (upsert - handles continued sessions)
	if err := sessionRepo.Create(ctx, session); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Clear any existing quality data (stale after session continuation)
	_ = qualityRepo.Delete(ctx, session.ID)

	// Save metrics
	if err := metricsRepo.Create(ctx, parsed.Metrics); err != nil {
		return fmt.Errorf("failed to create session metrics: %w", err)
	}

	// Save tools
	if len(parsed.Tools) > 0 {
		if err := toolRepo.CreateBatch(ctx, parsed.Tools); err != nil {
			return fmt.Errorf("failed to create session tools: %w", err)
		}
	}

	// Save files
	if len(parsed.Files) > 0 {
		if err := fileRepo.CreateBatch(ctx, parsed.Files); err != nil {
			return fmt.Errorf("failed to create session files: %w", err)
		}
	}

	// Save commands
	if len(parsed.Commands) > 0 {
		if err := commandRepo.CreateBatch(ctx, parsed.Commands); err != nil {
			return fmt.Errorf("failed to create session commands: %w", err)
		}
	}

	// Calculate sub-agent cost estimates and save
	if len(parsed.Subagents) > 0 {
		for _, sa := range parsed.Subagents {
			// Try to find pricing for sub-agent model
			var saPricing *domain.ModelPricing
			if sa.Model != nil {
				// Try alias lookup (e.g., "haiku" -> model ID)
				if modelID := resolveModelAlias(*sa.Model); modelID != "" {
					saPricing, _ = pricingRepo.GetByID(ctx, modelID)
				}
			}
			if saPricing == nil {
				saPricing = pricing // Fall back to session's default pricing
			}
			if saPricing != nil {
				cost := saPricing.CalculateCost(sa.TokenInput, sa.TokenOutput, sa.TokenCacheRead, sa.TokenCacheWrite)
				sa.CostEstimateUSD = &cost
			}
		}
		if err := subagentRepo.CreateBatch(ctx, parsed.Subagents); err != nil {
			return fmt.Errorf("failed to create session subagents: %w", err)
		}
	}

	// Sync to remote if enabled (only for real Turso connection)
	if tursoDB != nil {
		if err := tursoDB.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: sync failed: %v\n", err)
		}
	}

	// Output success message (goes to stdout, visible in hook output)
	fmt.Printf("Session %s recorded: %d input tokens, %d output tokens",
		hookInput.SessionID[:8],
		parsed.Metrics.TokenInput,
		parsed.Metrics.TokenOutput,
	)
	if costEstimate != nil {
		fmt.Printf(", $%.4f estimated cost", *costEstimate)
	}
	if len(parsed.Subagents) > 0 {
		fmt.Printf(", %d sub-agents", len(parsed.Subagents))
	}
	fmt.Println()

	return nil
}

// resolveModelAlias maps short model aliases to full model IDs for pricing lookup.
func resolveModelAlias(alias string) string {
	aliases := map[string]string{
		"haiku":  "claude-haiku-4-5-20251001",
		"sonnet": "claude-sonnet-4-5-20250929",
		"opus":   "claude-opus-4-6-20260206",
	}
	if id, ok := aliases[alias]; ok {
		return id
	}
	// If it's already a full model ID, return as-is
	return alias
}
