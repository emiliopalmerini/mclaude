package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/adapters/otel"
	"github.com/emiliopalmerini/mclaude/internal/adapters/storage"
	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/parser"
	"github.com/emiliopalmerini/mclaude/internal/ports"
)

var recordBackgroundFile string
var recordSync bool

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
              "command": "mclaude record"
            }
          ]
        }
      ]
    }
  }`,
	RunE: runRecord,
}

func init() {
	recordCmd.Flags().StringVar(&recordBackgroundFile, "background", "", "process hook input from file (internal use)")
	recordCmd.Flags().MarkHidden("background")
	recordCmd.Flags().BoolVar(&recordSync, "sync", false, "process synchronously (for debugging)")
}

func runRecord(cmd *cobra.Command, args []string) error {
	// If --background flag is set, process from file
	if recordBackgroundFile != "" {
		return processRecordBackground(recordBackgroundFile)
	}

	// Read hook input from stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	var hookInput domain.HookInput
	if err := json.Unmarshal(input, &hookInput); err != nil {
		return fmt.Errorf("failed to parse hook input: %w", err)
	}

	// Process synchronously if --sync flag is set
	if recordSync {
		return processRecordInput(&hookInput)
	}

	// Write input to temp file for background processing
	tempDir := os.TempDir()
	tempFile := filepath.Join(tempDir, fmt.Sprintf("mclaude-record-%s.json", hookInput.SessionID))
	if err := os.WriteFile(tempFile, input, 0600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Spawn background process
	executable, err := os.Executable()
	if err != nil {
		// Fallback to synchronous if we can't find executable
		os.Remove(tempFile)
		return processRecordInput(&hookInput)
	}

	bgCmd := exec.Command(executable, "record", "--background", tempFile)
	bgCmd.Stdout = nil
	bgCmd.Stderr = nil
	bgCmd.Stdin = nil

	if err := bgCmd.Start(); err != nil {
		// Fallback to synchronous if spawn fails
		os.Remove(tempFile)
		return processRecordInput(&hookInput)
	}

	// Detach from child process
	bgCmd.Process.Release()

	return nil
}

func processRecordBackground(inputFile string) error {
	// Read hook input from file
	input, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Clean up temp file
	defer os.Remove(inputFile)

	var hookInput domain.HookInput
	if err := json.Unmarshal(input, &hookInput); err != nil {
		return fmt.Errorf("failed to parse hook input: %w", err)
	}

	return processRecordInput(&hookInput)
}

func processRecordInput(hookInput *domain.HookInput) error {
	ctx := context.Background()

	// Connect to database
	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Initialize repositories (pass embedded *sql.DB to constructors)
	projectRepo := turso.NewProjectRepository(db.DB)
	experimentRepo := turso.NewExperimentRepository(db.DB)
	sessionRepo := turso.NewSessionRepository(db.DB)
	metricsRepo := turso.NewSessionMetricsRepository(db.DB)
	toolRepo := turso.NewSessionToolRepository(db.DB)
	fileRepo := turso.NewSessionFileRepository(db.DB)
	commandRepo := turso.NewSessionCommandRepository(db.DB)
	pricingRepo := turso.NewPricingRepository(db.DB)
	qualityRepo := turso.NewSessionQualityRepository(db.DB)
	planConfigRepo := turso.NewPlanConfigRepository(db.DB)

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

	// Sync to remote if enabled
	if err := db.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: sync failed: %v\n", err)
	}

	// Export to OTEL if configured
	exportOTELMetrics(ctx, session, project, activeExperiment, parsed, costEstimate)

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

// exportOTELMetrics exports enriched session metrics to OTEL Collector.
func exportOTELMetrics(
	ctx context.Context,
	session *domain.Session,
	project *domain.Project,
	experiment *domain.Experiment,
	parsed *parser.ParsedTranscript,
	costEstimate *float64,
) {
	cfg := otel.LoadConfig()
	if !cfg.Enabled || cfg.Endpoint == "" {
		return
	}

	exporter, err := otel.NewExporter(ctx, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: OTEL exporter init failed: %v\n", err)
		return
	}
	defer exporter.Close(ctx)

	metrics := &ports.EnrichedMetrics{
		SessionID:       session.ID,
		ProjectID:       project.ID,
		ProjectName:     project.Name,
		TokenInput:      parsed.Metrics.TokenInput,
		TokenOutput:     parsed.Metrics.TokenOutput,
		TokenCacheRead:  parsed.Metrics.TokenCacheRead,
		TokenCacheWrite: parsed.Metrics.TokenCacheWrite,
		TurnCount:       parsed.Metrics.TurnCount,
		ErrorCount:      parsed.Metrics.ErrorCount,
		ExitReason:      session.ExitReason,
	}

	if experiment != nil {
		metrics.ExperimentID = &experiment.ID
		metrics.ExperimentName = &experiment.Name
	}

	if costEstimate != nil {
		metrics.CostEstimateUSD = *costEstimate
	}

	if session.DurationSeconds != nil {
		metrics.DurationSeconds = *session.DurationSeconds
	}

	if session.StartedAt != nil {
		metrics.StartedAt = *session.StartedAt
	}
	if session.EndedAt != nil {
		metrics.EndedAt = *session.EndedAt
	}

	if err := exporter.ExportSessionMetrics(ctx, metrics); err != nil {
		fmt.Fprintf(os.Stderr, "warning: OTEL export failed: %v\n", err)
	}
}
