package turso

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/util"
	"github.com/emiliopalmerini/mclaude/sqlc/generated"
)

type StatsRepository struct {
	queries *sqlc.Queries
}

func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{queries: sqlc.New(db)}
}

func (r *StatsRepository) GetAggregate(ctx context.Context, since string) (*domain.AggregateStats, error) {
	row, err := r.queries.GetAggregateStats(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get aggregate stats: %w", err)
	}
	return &domain.AggregateStats{
		SessionCount:           row.SessionCount,
		TotalUserMessages:      util.ToInt64(row.TotalUserMessages),
		TotalAssistantMessages: util.ToInt64(row.TotalAssistantMessages),
		TotalTurns:             util.ToInt64(row.TotalTurns),
		TotalTokenInput:        util.ToInt64(row.TotalTokenInput),
		TotalTokenOutput:       util.ToInt64(row.TotalTokenOutput),
		TotalTokenCacheRead:    util.ToInt64(row.TotalTokenCacheRead),
		TotalTokenCacheWrite:   util.ToInt64(row.TotalTokenCacheWrite),
		TotalCostUsd:           util.ToFloat64(row.TotalCostUsd),
		TotalErrors:            util.ToInt64(row.TotalErrors),
	}, nil
}

func (r *StatsRepository) GetAggregateByExperiment(ctx context.Context, experimentID string, since string) (*domain.AggregateStats, error) {
	row, err := r.queries.GetAggregateStatsByExperiment(ctx, sqlc.GetAggregateStatsByExperimentParams{
		ExperimentID: util.NullString(experimentID),
		CreatedAt:    since,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment stats: %w", err)
	}
	return &domain.AggregateStats{
		SessionCount:           row.SessionCount,
		TotalUserMessages:      util.ToInt64(row.TotalUserMessages),
		TotalAssistantMessages: util.ToInt64(row.TotalAssistantMessages),
		TotalTurns:             util.ToInt64(row.TotalTurns),
		TotalTokenInput:        util.ToInt64(row.TotalTokenInput),
		TotalTokenOutput:       util.ToInt64(row.TotalTokenOutput),
		TotalTokenCacheRead:    util.ToInt64(row.TotalTokenCacheRead),
		TotalTokenCacheWrite:   util.ToInt64(row.TotalTokenCacheWrite),
		TotalCostUsd:           util.ToFloat64(row.TotalCostUsd),
		TotalErrors:            util.ToInt64(row.TotalErrors),
	}, nil
}

func (r *StatsRepository) GetAggregateByProject(ctx context.Context, projectID string, since string) (*domain.AggregateStats, error) {
	row, err := r.queries.GetAggregateStatsByProject(ctx, sqlc.GetAggregateStatsByProjectParams{
		ProjectID: projectID,
		CreatedAt: since,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get project stats: %w", err)
	}
	return &domain.AggregateStats{
		SessionCount:           row.SessionCount,
		TotalUserMessages:      util.ToInt64(row.TotalUserMessages),
		TotalAssistantMessages: util.ToInt64(row.TotalAssistantMessages),
		TotalTurns:             util.ToInt64(row.TotalTurns),
		TotalTokenInput:        util.ToInt64(row.TotalTokenInput),
		TotalTokenOutput:       util.ToInt64(row.TotalTokenOutput),
		TotalTokenCacheRead:    util.ToInt64(row.TotalTokenCacheRead),
		TotalTokenCacheWrite:   util.ToInt64(row.TotalTokenCacheWrite),
		TotalCostUsd:           util.ToFloat64(row.TotalCostUsd),
		TotalErrors:            util.ToInt64(row.TotalErrors),
	}, nil
}

func (r *StatsRepository) GetTopTools(ctx context.Context, since string, limit int) ([]domain.ToolUsageStats, error) {
	rows, err := r.queries.GetTopToolsUsage(ctx, sqlc.GetTopToolsUsageParams{
		CreatedAt: since,
		Limit:     int64(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get top tools: %w", err)
	}
	tools := make([]domain.ToolUsageStats, len(rows))
	for i, row := range rows {
		tools[i] = domain.ToolUsageStats{
			ToolName:         row.ToolName,
			TotalInvocations: int64(util.ToFloat64(row.TotalInvocations)),
			TotalErrors:      int64(util.ToFloat64(row.TotalErrors)),
		}
	}
	return tools, nil
}

func (r *StatsRepository) GetTotalToolCallsByExperiment(ctx context.Context, experimentID string) (int64, error) {
	result, err := r.queries.GetTotalToolCallsByExperiment(ctx, util.NullString(experimentID))
	if err != nil {
		return 0, fmt.Errorf("failed to get tool call count: %w", err)
	}
	return util.ToInt64(result), nil
}

func (r *StatsRepository) GetAllExperimentStats(ctx context.Context) ([]domain.ExperimentStats, error) {
	rows, err := r.queries.GetStatsForAllExperiments(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment stats: %w", err)
	}
	stats := make([]domain.ExperimentStats, len(rows))
	for i, row := range rows {
		stats[i] = domain.ExperimentStats{
			ExperimentID:   row.ExperimentID,
			ExperimentName: row.ExperimentName,
			AggregateStats: domain.AggregateStats{
				SessionCount:           row.SessionCount,
				TotalUserMessages:      util.ToInt64(row.TotalUserMessages),
				TotalAssistantMessages: util.ToInt64(row.TotalAssistantMessages),
				TotalTurns:             util.ToInt64(row.TotalTurns),
				TotalTokenInput:        util.ToInt64(row.TotalTokenInput),
				TotalTokenOutput:       util.ToInt64(row.TotalTokenOutput),
				TotalTokenCacheRead:    util.ToInt64(row.TotalTokenCacheRead),
				TotalTokenCacheWrite:   util.ToInt64(row.TotalTokenCacheWrite),
				TotalCostUsd:           util.ToFloat64(row.TotalCostUsd),
				TotalErrors:            util.ToInt64(row.TotalErrors),
			},
		}
	}
	return stats, nil
}
