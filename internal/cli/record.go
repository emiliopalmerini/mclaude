package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/claude-watcher/internal/adapters/storage"
	"github.com/emiliopalmerini/claude-watcher/internal/adapters/turso"
	"github.com/emiliopalmerini/claude-watcher/internal/domain"
	"github.com/emiliopalmerini/claude-watcher/internal/parser"
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record session data from Claude Code hook",
	Long: `Reads session data from stdin (Claude Code SessionEnd hook),
parses the transcript, and saves all data to the database.

This command is designed to be called from a Claude Code hook:

  {
    "hooks": {
      "SessionEnd": [
        {
          "hooks": [
            {
              "type": "command",
              "command": "claude-watcher record"
            }
          ]
        }
      ]
    }
  }`,
	RunE: runRecord,
}

func runRecord(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Read hook input from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	var hookInput domain.HookInput
	if err := json.Unmarshal(input, &hookInput); err != nil {
		return fmt.Errorf("failed to parse hook input: %w", err)
	}

	// Connect to database
	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repositories
	projectRepo := turso.NewProjectRepository(db)
	experimentRepo := turso.NewExperimentRepository(db)
	sessionRepo := turso.NewSessionRepository(db)
	metricsRepo := turso.NewSessionMetricsRepository(db)
	toolRepo := turso.NewSessionToolRepository(db)
	fileRepo := turso.NewSessionFileRepository(db)
	commandRepo := turso.NewSessionCommandRepository(db)
	pricingRepo := turso.NewPricingRepository(db)

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

	// Store transcript copy
	storedPath, err := transcriptStorage.Store(ctx, hookInput.SessionID, hookInput.TranscriptPath)
	if err != nil {
		// Log but don't fail - transcript storage is not critical
		fmt.Fprintf(os.Stderr, "warning: failed to store transcript copy: %v\n", err)
	}

	// Calculate cost estimate
	var costEstimate *float64
	defaultPricing, err := pricingRepo.GetDefault(ctx)
	if err == nil && defaultPricing != nil {
		cost := defaultPricing.CalculateCost(
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
		ID:             hookInput.SessionID,
		ProjectID:      project.ID,
		TranscriptPath: hookInput.TranscriptPath,
		Cwd:            hookInput.Cwd,
		PermissionMode: hookInput.PermissionMode,
		ExitReason:     hookInput.Reason,
		StartedAt:      parsed.StartedAt,
		EndedAt:        parsed.EndedAt,
		DurationSeconds: durationSeconds,
		CreatedAt:      time.Now().UTC(),
	}

	if storedPath != "" {
		session.TranscriptStoredPath = &storedPath
	}

	if activeExperiment != nil {
		session.ExperimentID = &activeExperiment.ID
	}

	// Save session
	if err := sessionRepo.Create(ctx, session); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

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

	// Output success message (goes to stdout, visible in hook output)
	fmt.Printf("Session %s recorded: %d input tokens, %d output tokens",
		hookInput.SessionID[:8],
		parsed.Metrics.TokenInput,
		parsed.Metrics.TokenOutput,
	)
	if costEstimate != nil {
		fmt.Printf(", $%.4f estimated cost", *costEstimate)
	}
	fmt.Println()

	return nil
}
