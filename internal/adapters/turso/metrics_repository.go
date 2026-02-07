package turso

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/sqlc/generated"
)

type SessionMetricsRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewSessionMetricsRepository(db *sql.DB) *SessionMetricsRepository {
	return &SessionMetricsRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *SessionMetricsRepository) Create(ctx context.Context, metrics *domain.SessionMetrics) error {
	var costEstimate sql.NullFloat64
	if metrics.CostEstimateUSD != nil {
		costEstimate = sql.NullFloat64{Float64: *metrics.CostEstimateUSD, Valid: true}
	}

	var modelID sql.NullString
	if metrics.ModelID != nil {
		modelID = sql.NullString{String: *metrics.ModelID, Valid: true}
	}

	return r.queries.CreateSessionMetrics(ctx, sqlc.CreateSessionMetricsParams{
		SessionID:             metrics.SessionID,
		ModelID:               modelID,
		MessageCountUser:      metrics.MessageCountUser,
		MessageCountAssistant: metrics.MessageCountAssistant,
		TurnCount:             metrics.TurnCount,
		TokenInput:            metrics.TokenInput,
		TokenOutput:           metrics.TokenOutput,
		TokenCacheRead:        metrics.TokenCacheRead,
		TokenCacheWrite:       metrics.TokenCacheWrite,
		CostEstimateUsd:       costEstimate,
		ErrorCount:            metrics.ErrorCount,
	})
}

func (r *SessionMetricsRepository) GetBySessionID(ctx context.Context, sessionID string) (*domain.SessionMetrics, error) {
	row, err := r.queries.GetSessionMetricsBySessionID(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session metrics: %w", err)
	}

	var costEstimate *float64
	if row.CostEstimateUsd.Valid {
		costEstimate = &row.CostEstimateUsd.Float64
	}

	var modelID *string
	if row.ModelID.Valid {
		modelID = &row.ModelID.String
	}

	return &domain.SessionMetrics{
		SessionID:             row.SessionID,
		ModelID:               modelID,
		MessageCountUser:      row.MessageCountUser,
		MessageCountAssistant: row.MessageCountAssistant,
		TurnCount:             row.TurnCount,
		TokenInput:            row.TokenInput,
		TokenOutput:           row.TokenOutput,
		TokenCacheRead:        row.TokenCacheRead,
		TokenCacheWrite:       row.TokenCacheWrite,
		CostEstimateUSD:       costEstimate,
		ErrorCount:            row.ErrorCount,
	}, nil
}

type SessionToolRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewSessionToolRepository(db *sql.DB) *SessionToolRepository {
	return &SessionToolRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *SessionToolRepository) CreateBatch(ctx context.Context, tools []*domain.SessionTool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	for _, tool := range tools {
		var totalDurationMs sql.NullInt64
		if tool.TotalDurationMs != nil {
			totalDurationMs = sql.NullInt64{Int64: *tool.TotalDurationMs, Valid: true}
		}

		err := qtx.CreateSessionTool(ctx, sqlc.CreateSessionToolParams{
			SessionID:       tool.SessionID,
			ToolName:        tool.ToolName,
			InvocationCount: tool.InvocationCount,
			TotalDurationMs: totalDurationMs,
			ErrorCount:      tool.ErrorCount,
		})
		if err != nil {
			return fmt.Errorf("failed to create session tool %s: %w", tool.ToolName, err)
		}
	}
	return tx.Commit()
}

func (r *SessionToolRepository) ListBySessionID(ctx context.Context, sessionID string) ([]*domain.SessionTool, error) {
	rows, err := r.queries.ListSessionToolsBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list session tools: %w", err)
	}

	tools := make([]*domain.SessionTool, len(rows))
	for i, row := range rows {
		var totalDurationMs *int64
		if row.TotalDurationMs.Valid {
			totalDurationMs = &row.TotalDurationMs.Int64
		}
		tools[i] = &domain.SessionTool{
			ID:              row.ID,
			SessionID:       row.SessionID,
			ToolName:        row.ToolName,
			InvocationCount: row.InvocationCount,
			TotalDurationMs: totalDurationMs,
			ErrorCount:      row.ErrorCount,
		}
	}
	return tools, nil
}

type SessionFileRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewSessionFileRepository(db *sql.DB) *SessionFileRepository {
	return &SessionFileRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *SessionFileRepository) CreateBatch(ctx context.Context, files []*domain.SessionFile) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	for _, file := range files {
		err := qtx.CreateSessionFile(ctx, sqlc.CreateSessionFileParams{
			SessionID:      file.SessionID,
			FilePath:       file.FilePath,
			Operation:      file.Operation,
			OperationCount: file.OperationCount,
		})
		if err != nil {
			return fmt.Errorf("failed to create session file %s: %w", file.FilePath, err)
		}
	}
	return tx.Commit()
}

func (r *SessionFileRepository) ListBySessionID(ctx context.Context, sessionID string) ([]*domain.SessionFile, error) {
	rows, err := r.queries.ListSessionFilesBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list session files: %w", err)
	}

	files := make([]*domain.SessionFile, len(rows))
	for i, row := range rows {
		files[i] = &domain.SessionFile{
			ID:             row.ID,
			SessionID:      row.SessionID,
			FilePath:       row.FilePath,
			Operation:      row.Operation,
			OperationCount: row.OperationCount,
		}
	}
	return files, nil
}

type SessionCommandRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewSessionCommandRepository(db *sql.DB) *SessionCommandRepository {
	return &SessionCommandRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *SessionCommandRepository) CreateBatch(ctx context.Context, commands []*domain.SessionCommand) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	for _, cmd := range commands {
		var exitCode sql.NullInt64
		if cmd.ExitCode != nil {
			exitCode = sql.NullInt64{Int64: int64(*cmd.ExitCode), Valid: true}
		}

		var executedAt sql.NullString
		if cmd.ExecutedAt != nil {
			executedAt = sql.NullString{String: cmd.ExecutedAt.Format(time.RFC3339), Valid: true}
		}

		err := qtx.CreateSessionCommand(ctx, sqlc.CreateSessionCommandParams{
			SessionID:  cmd.SessionID,
			Command:    cmd.Command,
			ExitCode:   exitCode,
			ExecutedAt: executedAt,
		})
		if err != nil {
			return fmt.Errorf("failed to create session command: %w", err)
		}
	}
	return tx.Commit()
}

func (r *SessionCommandRepository) ListBySessionID(ctx context.Context, sessionID string) ([]*domain.SessionCommand, error) {
	rows, err := r.queries.ListSessionCommandsBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list session commands: %w", err)
	}

	commands := make([]*domain.SessionCommand, len(rows))
	for i, row := range rows {
		var exitCode *int
		if row.ExitCode.Valid {
			ec := int(row.ExitCode.Int64)
			exitCode = &ec
		}

		var executedAt *time.Time
		if row.ExecutedAt.Valid {
			t, _ := time.Parse(time.RFC3339, row.ExecutedAt.String)
			executedAt = &t
		}

		commands[i] = &domain.SessionCommand{
			ID:         row.ID,
			SessionID:  row.SessionID,
			Command:    row.Command,
			ExitCode:   exitCode,
			ExecutedAt: executedAt,
		}
	}
	return commands, nil
}

type SessionSubagentRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewSessionSubagentRepository(db *sql.DB) *SessionSubagentRepository {
	return &SessionSubagentRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *SessionSubagentRepository) CreateBatch(ctx context.Context, subagents []*domain.SessionSubagent) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)
	for _, sa := range subagents {
		var description sql.NullString
		if sa.Description != nil {
			description = sql.NullString{String: *sa.Description, Valid: true}
		}

		var model sql.NullString
		if sa.Model != nil {
			model = sql.NullString{String: *sa.Model, Valid: true}
		}

		var totalDurationMs sql.NullInt64
		if sa.TotalDurationMs != nil {
			totalDurationMs = sql.NullInt64{Int64: *sa.TotalDurationMs, Valid: true}
		}

		var costEstimate sql.NullFloat64
		if sa.CostEstimateUSD != nil {
			costEstimate = sql.NullFloat64{Float64: *sa.CostEstimateUSD, Valid: true}
		}

		err := qtx.CreateSessionSubagent(ctx, sqlc.CreateSessionSubagentParams{
			SessionID:       sa.SessionID,
			AgentType:       sa.AgentType,
			AgentKind:       sa.AgentKind,
			Description:     description,
			Model:           model,
			TotalTokens:     sa.TotalTokens,
			TokenInput:      sa.TokenInput,
			TokenOutput:     sa.TokenOutput,
			TokenCacheRead:  sa.TokenCacheRead,
			TokenCacheWrite: sa.TokenCacheWrite,
			TotalDurationMs: totalDurationMs,
			ToolUseCount:    sa.ToolUseCount,
			CostEstimateUsd: costEstimate,
		})
		if err != nil {
			return fmt.Errorf("failed to create session subagent %s: %w", sa.AgentType, err)
		}
	}
	return tx.Commit()
}

func (r *SessionSubagentRepository) ListBySessionID(ctx context.Context, sessionID string) ([]*domain.SessionSubagent, error) {
	rows, err := r.queries.ListSessionSubagentsBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list session subagents: %w", err)
	}

	subagents := make([]*domain.SessionSubagent, len(rows))
	for i, row := range rows {
		var description *string
		if row.Description.Valid {
			description = &row.Description.String
		}

		var model *string
		if row.Model.Valid {
			model = &row.Model.String
		}

		var totalDurationMs *int64
		if row.TotalDurationMs.Valid {
			totalDurationMs = &row.TotalDurationMs.Int64
		}

		var costEstimate *float64
		if row.CostEstimateUsd.Valid {
			costEstimate = &row.CostEstimateUsd.Float64
		}

		subagents[i] = &domain.SessionSubagent{
			ID:              row.ID,
			SessionID:       row.SessionID,
			AgentType:       row.AgentType,
			AgentKind:       row.AgentKind,
			Description:     description,
			Model:           model,
			TotalTokens:     row.TotalTokens,
			TokenInput:      row.TokenInput,
			TokenOutput:     row.TokenOutput,
			TokenCacheRead:  row.TokenCacheRead,
			TokenCacheWrite: row.TokenCacheWrite,
			TotalDurationMs: totalDurationMs,
			ToolUseCount:    row.ToolUseCount,
			CostEstimateUSD: costEstimate,
		}
	}
	return subagents, nil
}
