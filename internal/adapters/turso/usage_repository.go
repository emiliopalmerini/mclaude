package turso

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

type UsageMetricsRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewUsageMetricsRepository(db *sql.DB) *UsageMetricsRepository {
	return &UsageMetricsRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *UsageMetricsRepository) Create(ctx context.Context, metric *domain.UsageMetric) error {
	params := sqlc.CreateUsageMetricParams{
		MetricName: metric.MetricName,
		Value:      metric.Value,
		RecordedAt: metric.RecordedAt.Format(time.RFC3339),
	}
	if metric.Attributes != nil {
		params.Attributes = sql.NullString{String: *metric.Attributes, Valid: true}
	}
	return r.queries.CreateUsageMetric(ctx, params)
}

func (r *UsageMetricsRepository) GetDailySummary(ctx context.Context) (*domain.UsageSummary, error) {
	row, err := r.queries.GetDailyUsageSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily usage summary: %w", err)
	}
	return &domain.UsageSummary{
		TotalTokens: row.TotalTokens,
		TotalCost:   row.TotalCost,
	}, nil
}

func (r *UsageMetricsRepository) GetWeeklySummary(ctx context.Context) (*domain.UsageSummary, error) {
	row, err := r.queries.GetWeeklyUsageSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get weekly usage summary: %w", err)
	}
	return &domain.UsageSummary{
		TotalTokens: row.TotalTokens,
		TotalCost:   row.TotalCost,
	}, nil
}

func (r *UsageMetricsRepository) DeleteBefore(ctx context.Context, before string) (int64, error) {
	return r.queries.DeleteUsageMetricsBefore(ctx, before)
}

func (r *UsageMetricsRepository) GetRollingWindowSummary(ctx context.Context, windowHours int) (*domain.UsageSummary, error) {
	// SQLite datetime modifier needs negative hours
	hoursParam := fmt.Sprintf("-%d", windowHours)
	row, err := r.queries.GetRollingWindowUsage(ctx, sql.NullString{String: hoursParam, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get rolling window usage: %w", err)
	}
	return &domain.UsageSummary{
		TotalTokens: row.TotalTokens,
		TotalCost:   row.TotalCost,
	}, nil
}

type UsageLimitsRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewUsageLimitsRepository(db *sql.DB) *UsageLimitsRepository {
	return &UsageLimitsRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *UsageLimitsRepository) Upsert(ctx context.Context, limit *domain.UsageLimit) error {
	enabled := int64(0)
	if limit.Enabled {
		enabled = 1
	}
	return r.queries.CreateUsageLimit(ctx, sqlc.CreateUsageLimitParams{
		ID:            limit.ID,
		LimitValue:    limit.LimitValue,
		WarnThreshold: sql.NullFloat64{Float64: limit.WarnThreshold, Valid: true},
		Enabled:       enabled,
	})
}

func (r *UsageLimitsRepository) Get(ctx context.Context, id string) (*domain.UsageLimit, error) {
	row, err := r.queries.GetUsageLimit(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get usage limit: %w", err)
	}
	return limitFromRow(row), nil
}

func (r *UsageLimitsRepository) List(ctx context.Context) ([]*domain.UsageLimit, error) {
	rows, err := r.queries.ListUsageLimits(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list usage limits: %w", err)
	}
	limits := make([]*domain.UsageLimit, len(rows))
	for i, row := range rows {
		limits[i] = limitFromRow(row)
	}
	return limits, nil
}

func (r *UsageLimitsRepository) Delete(ctx context.Context, id string) error {
	return r.queries.DeleteUsageLimit(ctx, id)
}

func limitFromRow(row sqlc.UsageLimit) *domain.UsageLimit {
	limit := &domain.UsageLimit{
		ID:            row.ID,
		LimitValue:    row.LimitValue,
		WarnThreshold: 0.8,
		Enabled:       row.Enabled == 1,
		CreatedAt:     parseTime(row.CreatedAt),
		UpdatedAt:     parseTime(row.UpdatedAt),
	}
	if row.WarnThreshold.Valid {
		limit.WarnThreshold = row.WarnThreshold.Float64
	}
	return limit
}

type PlanConfigRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewPlanConfigRepository(db *sql.DB) *PlanConfigRepository {
	return &PlanConfigRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *PlanConfigRepository) Upsert(ctx context.Context, config *domain.PlanConfig) error {
	var learnedLimit sql.NullFloat64
	var learnedAt sql.NullString

	if config.LearnedTokenLimit != nil {
		learnedLimit = sql.NullFloat64{Float64: *config.LearnedTokenLimit, Valid: true}
	}
	if config.LearnedAt != nil {
		learnedAt = sql.NullString{String: config.LearnedAt.Format(time.RFC3339), Valid: true}
	}

	return r.queries.UpsertPlanConfig(ctx, sqlc.UpsertPlanConfigParams{
		PlanType:          config.PlanType,
		WindowHours:       int64(config.WindowHours),
		LearnedTokenLimit: learnedLimit,
		LearnedAt:         learnedAt,
	})
}

func (r *PlanConfigRepository) Get(ctx context.Context) (*domain.PlanConfig, error) {
	row, err := r.queries.GetPlanConfig(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get plan config: %w", err)
	}
	return planConfigFromRow(row), nil
}

func (r *PlanConfigRepository) UpdateLearnedLimit(ctx context.Context, limit float64) error {
	return r.queries.UpdateLearnedLimit(ctx, sql.NullFloat64{Float64: limit, Valid: true})
}

func planConfigFromRow(row sqlc.PlanConfig) *domain.PlanConfig {
	config := &domain.PlanConfig{
		PlanType:    row.PlanType,
		WindowHours: int(row.WindowHours),
		CreatedAt:   parseTime(row.CreatedAt),
		UpdatedAt:   parseTime(row.UpdatedAt),
	}
	if row.LearnedTokenLimit.Valid {
		config.LearnedTokenLimit = &row.LearnedTokenLimit.Float64
	}
	if row.LearnedAt.Valid {
		t, _ := time.Parse(time.RFC3339, row.LearnedAt.String)
		config.LearnedAt = &t
	}
	return config
}
