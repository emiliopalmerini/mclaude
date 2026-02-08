package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/parser"
)

func handleSubagentStart(event *domain.SubagentStartInput) error {
	sqlDB, _, closeDB, err := hookDB()
	if err != nil {
		return err
	}
	defer closeDB()

	ctx := context.Background()
	now := time.Now().UTC().Format(time.RFC3339)

	_, err = sqlDB.ExecContext(ctx,
		"INSERT OR REPLACE INTO hook_subagent_tracking (agent_id, session_id, agent_type, started_at) VALUES (?, ?, ?, ?)",
		event.AgentID, event.SessionID, event.AgentType, now,
	)
	if err != nil {
		return fmt.Errorf("failed to insert subagent tracking: %w", err)
	}

	return nil
}

func handleSubagentStop(event *domain.SubagentStopInput) error {
	sqlDB, tursoDB, closeDB, err := hookDB()
	if err != nil {
		return err
	}
	defer func() { syncAndClose(tursoDB, closeDB) }()

	ctx := context.Background()

	// Look up tracking row
	var sessionID, agentType, startedAt string
	err = sqlDB.QueryRowContext(ctx,
		"SELECT session_id, agent_type, started_at FROM hook_subagent_tracking WHERE agent_id = ?",
		event.AgentID,
	).Scan(&sessionID, &agentType, &startedAt)
	if err != nil {
		// No tracking row found — graceful handling
		fmt.Fprintf(os.Stderr, "warning: no tracking row for agent %s, skipping\n", event.AgentID)
		return nil
	}

	// Always clean up the tracking row
	defer func() {
		_, _ = sqlDB.ExecContext(ctx,
			"DELETE FROM hook_subagent_tracking WHERE agent_id = ?",
			event.AgentID,
		)
	}()

	// Parse the sub-agent transcript
	if event.AgentTranscriptPath == "" {
		fmt.Fprintf(os.Stderr, "warning: no transcript path for agent %s\n", event.AgentID)
		return nil
	}

	parsed, err := parser.ParseTranscript(sessionID, event.AgentTranscriptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to parse subagent transcript: %v\n", err)
		return nil
	}

	// Calculate duration from tracking start time
	var durationMs *int64
	startTime, err := time.Parse(time.RFC3339, startedAt)
	if err == nil {
		dur := time.Since(startTime).Milliseconds()
		durationMs = &dur
	}

	// Calculate total tokens
	totalTokens := parsed.Metrics.TokenInput + parsed.Metrics.TokenOutput +
		parsed.Metrics.TokenCacheRead + parsed.Metrics.TokenCacheWrite

	// Calculate cost estimate
	pricingRepo := turso.NewPricingRepository(sqlDB)
	var costEstimate *float64

	// Try to get pricing for the model
	if parsed.ModelID != nil {
		if p, _ := pricingRepo.GetByID(ctx, *parsed.ModelID); p != nil {
			cost := p.CalculateCost(parsed.Metrics.TokenInput, parsed.Metrics.TokenOutput, parsed.Metrics.TokenCacheRead, parsed.Metrics.TokenCacheWrite)
			costEstimate = &cost
		}
	}
	if costEstimate == nil {
		if p, _ := pricingRepo.GetDefault(ctx); p != nil {
			cost := p.CalculateCost(parsed.Metrics.TokenInput, parsed.Metrics.TokenOutput, parsed.Metrics.TokenCacheRead, parsed.Metrics.TokenCacheWrite)
			costEstimate = &cost
		}
	}

	// Save to session_subagents
	subagentRepo := turso.NewSessionSubagentRepository(sqlDB)
	subagent := &domain.SessionSubagent{
		SessionID:       sessionID,
		AgentType:       agentType,
		AgentKind:       "hook",
		TotalTokens:     totalTokens,
		TokenInput:      parsed.Metrics.TokenInput,
		TokenOutput:     parsed.Metrics.TokenOutput,
		TokenCacheRead:  parsed.Metrics.TokenCacheRead,
		TokenCacheWrite: parsed.Metrics.TokenCacheWrite,
		TotalDurationMs: durationMs,
		ToolUseCount:    int64(len(parsed.Tools)),
		CostEstimateUSD: costEstimate,
	}

	if err := subagentRepo.CreateBatch(ctx, []*domain.SessionSubagent{subagent}); err != nil {
		return fmt.Errorf("failed to save subagent data: %w", err)
	}

	return nil
}
