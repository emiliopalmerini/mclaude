package ports

import (
	"context"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

type StatsRepository interface {
	GetAggregate(ctx context.Context, since string) (*domain.AggregateStats, error)
	GetAggregateByExperiment(ctx context.Context, experimentID string, since string) (*domain.AggregateStats, error)
	GetAggregateByProject(ctx context.Context, projectID string, since string) (*domain.AggregateStats, error)
	GetTopTools(ctx context.Context, since string, limit int) ([]domain.ToolUsageStats, error)
	GetAllExperimentStats(ctx context.Context) ([]domain.ExperimentStats, error)
	GetTotalToolCallsByExperiment(ctx context.Context, experimentID string) (int64, error)
}
