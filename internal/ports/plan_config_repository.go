package ports

import (
	"context"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

type PlanConfigRepository interface {
	Upsert(ctx context.Context, config *domain.PlanConfig) error
	Get(ctx context.Context) (*domain.PlanConfig, error)
	UpdateLearnedLimit(ctx context.Context, limit float64) error
	UpdateWeeklyLearnedLimit(ctx context.Context, limit float64) error
	GetRollingWindowSummary(ctx context.Context, windowHours int) (*domain.UsageSummary, error)
	GetWeeklyWindowSummary(ctx context.Context) (*domain.UsageSummary, error)
	UpdateWindowStartTime(ctx context.Context, t time.Time) error
	UpdateWeeklyWindowStartTime(ctx context.Context, t time.Time) error
	ResetWindowIfExpired(ctx context.Context, sessionStartTime time.Time) (bool, error)
	ResetWeeklyWindowIfExpired(ctx context.Context, sessionStartTime time.Time) (bool, error)
}
