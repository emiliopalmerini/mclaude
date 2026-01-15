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
		TotalCost:     toFloat64(weekRow.CostWeek),
		Tokens: analytics.TokenSummary{
			Input:    toInt64(weekRow.InputTokensWeek),
			Output:   toInt64(weekRow.OutputTokensWeek),
			Thinking: toInt64(weekRow.ThinkingTokensWeek),
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

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
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
	case int64:
		return float64(val)
	case nil:
		return 0
	default:
		return 0
	}
}
