package turso

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"claude-watcher/internal/analytics"
	"claude-watcher/internal/database/sqlc"
)

// Repository implements analytics.Repository using sqlc queries
type Repository struct {
	queries *sqlc.Queries
}

// NewRepository creates a new Turso repository for analytics
func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

// GetOverviewMetrics retrieves aggregate metrics for the dashboard overview
func (r *Repository) GetOverviewMetrics(ctx context.Context) (analytics.OverviewMetrics, error) {
	// Get week metrics (last 7 days) for the overview
	weekRow, err := r.queries.GetWeekMetrics(ctx)
	if err != nil {
		return analytics.OverviewMetrics{}, fmt.Errorf("get week metrics: %w", err)
	}

	// Get limit hits count (non-fatal if table doesn't exist)
	var limitHits int64
	limitHits, _ = r.queries.CountLimitEvents(ctx)

	// Get last limit hit timestamp (non-fatal if table doesn't exist)
	var lastLimitHit *time.Time
	lastLimitStr, err := r.queries.GetLastLimitEvent(ctx)
	if err == nil && lastLimitStr != "" {
		if t, parseErr := time.Parse(time.RFC3339, lastLimitStr); parseErr == nil {
			lastLimitHit = &t
		}
	}

	return analytics.OverviewMetrics{
		TotalSessions: int(weekRow.SessionsWeek),
		TotalCost:     weekRow.CostWeek,
		Tokens: analytics.TokenSummary{
			Input:    weekRow.InputTokensWeek,
			Output:   weekRow.OutputTokensWeek,
			Thinking: weekRow.ThinkingTokensWeek,
		},
		LimitHits:    int(limitHits),
		LastLimitHit: lastLimitHit,
	}, nil
}

// ListSessions returns paginated session summaries with total count
func (r *Repository) ListSessions(ctx context.Context, filter analytics.SessionFilter) ([]analytics.SessionSummary, int, error) {
	// Get total count
	total, err := r.queries.CountSessions(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count sessions: %w", err)
	}

	// Get sessions
	params := sqlc.ListSessionsParams{
		Limit:  int64(filter.Limit),
		Offset: int64(filter.Offset),
	}

	rows, err := r.queries.ListSessions(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("list sessions: %w", err)
	}

	sessions := make([]analytics.SessionSummary, 0, len(rows))
	for _, row := range rows {
		s := analytics.SessionSummary{
			SessionID:     row.SessionID,
			WorkingDir:    row.WorkingDirectory.String,
			Model:         row.Model.String,
			EstimatedCost: row.EstimatedCostUsd.Float64,
			ToolCalls:     int(row.ToolCalls.Int64),
		}

		if t, err := time.Parse(time.RFC3339, row.Timestamp); err == nil {
			s.Timestamp = t
		}

		sessions = append(sessions, s)
	}

	return sessions, int(total), nil
}

// GetSession retrieves detailed session information
func (r *Repository) GetSession(ctx context.Context, sessionID string) (analytics.SessionDetail, error) {
	row, err := r.queries.GetSessionByID(ctx, sessionID)
	if err != nil {
		return analytics.SessionDetail{}, fmt.Errorf("get session: %w", err)
	}

	detail := analytics.SessionDetail{
		SessionID:          row.SessionID,
		Hostname:           row.Hostname,
		ExitReason:         row.ExitReason.String,
		WorkingDirectory:   row.WorkingDirectory.String,
		GitBranch:          row.GitBranch.String,
		Model:              row.Model.String,
		ClaudeVersion:      row.ClaudeVersion.String,
		DurationSeconds:    int(row.DurationSeconds.Int64),
		EstimatedCost:      row.EstimatedCostUsd.Float64,
		UserPrompts:        int(row.UserPrompts.Int64),
		AssistantResponses: int(row.AssistantResponses.Int64),
		ToolCalls:          int(row.ToolCalls.Int64),
		ErrorsCount:        int(row.ErrorsCount.Int64),
		Summary:            row.Summary.String,
		Notes:              row.Notes.String,
		Tokens: analytics.TokenSummary{
			Input:      row.InputTokens.Int64,
			Output:     row.OutputTokens.Int64,
			Thinking:   row.ThinkingTokens.Int64,
			CacheRead:  row.CacheReadTokens.Int64,
			CacheWrite: row.CacheWriteTokens.Int64,
		},
	}

	if t, err := time.Parse(time.RFC3339, row.Timestamp); err == nil {
		detail.Timestamp = t
	}

	if row.Rating.Valid {
		rating := int(row.Rating.Int64)
		detail.Rating = &rating
	}
	if row.PromptSpecificity.Valid {
		ps := int(row.PromptSpecificity.Int64)
		detail.PromptSpecificity = &ps
	}
	if row.TaskCompletion.Valid {
		tc := int(row.TaskCompletion.Int64)
		detail.TaskCompletion = &tc
	}
	if row.CodeConfidence.Valid {
		cc := int(row.CodeConfidence.Int64)
		detail.CodeConfidence = &cc
	}

	return detail, nil
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case float32:
		return int64(val)
	case string:
		// Try to parse string as int
		var i int64
		fmt.Sscanf(val, "%d", &i)
		return i
	case nil:
		return 0
	default:
		return 0
	}
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case int:
		return float64(val)
	case string:
		// Try to parse string as float
		var f float64
		fmt.Sscanf(val, "%f", &f)
		return f
	case nil:
		return 0
	default:
		return 0
	}
}
