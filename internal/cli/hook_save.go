package cli

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/adapters/storage"
	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/parser"
	"github.com/emiliopalmerini/mclaude/internal/ports"
)

// saveSessionOpts controls which parts of session data to save.
type saveSessionOpts struct {
	// SkipTranscriptStorage skips copying the transcript to storage.
	// Used by Stop handler since the transcript file is still being written.
	SkipTranscriptStorage bool

	// ExitReason is the session exit reason (only set by SessionEnd).
	ExitReason string
}

// saveSessionData parses a transcript and saves session + metrics to the database.
// Shared by handleSessionEnd and handleStop.
func saveSessionData(ctx context.Context, sqlDB *sql.DB, sessionID, transcriptPath, cwd, permissionMode string, opts saveSessionOpts) error {
	projectRepo := turso.NewProjectRepository(sqlDB)
	experimentRepo := turso.NewExperimentRepository(sqlDB)
	sessionRepo := turso.NewSessionRepository(sqlDB)
	metricsRepo := turso.NewSessionMetricsRepository(sqlDB)
	toolRepo := turso.NewSessionToolRepository(sqlDB)
	fileRepo := turso.NewSessionFileRepository(sqlDB)
	commandRepo := turso.NewSessionCommandRepository(sqlDB)
	subagentRepo := turso.NewSessionSubagentRepository(sqlDB)
	pricingRepo := turso.NewPricingRepository(sqlDB)

	project, err := projectRepo.GetOrCreate(ctx, cwd)
	if err != nil {
		return fmt.Errorf("failed to get/create project: %w", err)
	}

	activeExperiment, err := experimentRepo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active experiment: %w", err)
	}

	parsed, err := parser.ParseTranscript(sessionID, transcriptPath)
	if err != nil {
		return fmt.Errorf("failed to parse transcript: %w", err)
	}

	// Store transcript copy (unless skipped)
	var storedPath string
	if !opts.SkipTranscriptStorage {
		transcriptStorage, err := storage.NewTranscriptStorage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to initialize transcript storage: %v\n", err)
		} else {
			storedPath, err = transcriptStorage.Store(ctx, sessionID, transcriptPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to store transcript copy: %v\n", err)
			}
		}
	}

	parsed.Metrics.ModelID = parsed.ModelID

	// Resolve pricing
	defaultPricing, _ := pricingRepo.GetDefault(ctx)
	pricing := resolvePricing(ctx, parsed.ModelID, pricingRepo, defaultPricing)

	var costEstimate *float64
	if pricing != nil {
		cost := pricing.CalculateCost(
			parsed.Metrics.TokenInput,
			parsed.Metrics.TokenOutput,
			parsed.Metrics.TokenCacheRead,
			parsed.Metrics.TokenCacheWrite,
		)
		costEstimate = &cost

		rates := pricing.ResolveRates(
			parsed.Metrics.TokenInput,
			parsed.Metrics.TokenOutput,
			parsed.Metrics.TokenCacheRead,
			parsed.Metrics.TokenCacheWrite,
		)
		parsed.Metrics.InputRate = &rates.Input
		parsed.Metrics.OutputRate = &rates.Output
		parsed.Metrics.CacheReadRate = rates.CacheRead
		parsed.Metrics.CacheWriteRate = rates.CacheWrite
	}
	parsed.Metrics.CostEstimateUSD = costEstimate

	// Calculate duration
	var durationSeconds *int64
	if parsed.StartedAt != nil && parsed.EndedAt != nil {
		dur := int64(parsed.EndedAt.Sub(*parsed.StartedAt).Seconds())
		durationSeconds = &dur
	}

	session := &domain.Session{
		ID:              sessionID,
		ProjectID:       project.ID,
		TranscriptPath:  transcriptPath,
		Cwd:             cwd,
		PermissionMode:  permissionMode,
		ExitReason:      opts.ExitReason,
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

	if err := sessionRepo.Create(ctx, session); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	if err := metricsRepo.Create(ctx, parsed.Metrics); err != nil {
		return fmt.Errorf("failed to create session metrics: %w", err)
	}

	if len(parsed.Tools) > 0 {
		if err := toolRepo.CreateBatch(ctx, parsed.Tools); err != nil {
			return fmt.Errorf("failed to create session tools: %w", err)
		}
	}

	if len(parsed.Files) > 0 {
		if err := fileRepo.CreateBatch(ctx, parsed.Files); err != nil {
			return fmt.Errorf("failed to create session files: %w", err)
		}
	}

	if len(parsed.Commands) > 0 {
		if err := commandRepo.CreateBatch(ctx, parsed.Commands); err != nil {
			return fmt.Errorf("failed to create session commands: %w", err)
		}
	}

	if len(parsed.Subagents) > 0 {
		for _, sa := range parsed.Subagents {
			if p := resolvePricing(ctx, sa.Model, pricingRepo, pricing); p != nil {
				cost := p.CalculateCost(sa.TokenInput, sa.TokenOutput, sa.TokenCacheRead, sa.TokenCacheWrite)
				sa.CostEstimateUSD = &cost
			}
		}
		if err := subagentRepo.CreateBatch(ctx, parsed.Subagents); err != nil {
			return fmt.Errorf("failed to create session subagents: %w", err)
		}
	}

	// Output success message
	fmt.Printf("Session %s recorded: %d input tokens, %d output tokens",
		sessionID[:min(8, len(sessionID))],
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

// resolvePricing looks up model-specific pricing, falling back to the provided default.
func resolvePricing(ctx context.Context, model *string, repo ports.PricingRepository, fallback *domain.ModelPricing) *domain.ModelPricing {
	if model != nil {
		if p, _ := repo.GetByID(ctx, resolveModelAlias(*model)); p != nil {
			return p
		}
	}
	return fallback
}
