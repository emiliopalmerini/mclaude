package ports

import (
	"context"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

type UsageMetricsRepository interface {
	Create(ctx context.Context, metric *domain.UsageMetric) error
	GetDailySummary(ctx context.Context) (*domain.UsageSummary, error)
	GetWeeklySummary(ctx context.Context) (*domain.UsageSummary, error)
	DeleteBefore(ctx context.Context, before string) (int64, error)
}

type UsageLimitsRepository interface {
	Upsert(ctx context.Context, limit *domain.UsageLimit) error
	Get(ctx context.Context, id string) (*domain.UsageLimit, error)
	List(ctx context.Context) ([]*domain.UsageLimit, error)
	Delete(ctx context.Context, id string) error
}
