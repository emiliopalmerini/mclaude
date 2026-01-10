package costs

import (
	"context"
	"database/sql"

	"claude-watcher/internal/database/sqlc"
)

type Repository interface {
	GetProjectMetrics(ctx context.Context, days string) ([]sqlc.GetProjectMetricsRow, error)
	GetModelEfficiency(ctx context.Context, days string) ([]sqlc.GetModelEfficiencyRow, error)
	GetCacheMetrics(ctx context.Context, days string) (sqlc.GetCacheMetricsRow, error)
	GetCacheMetricsDaily(ctx context.Context, days string) ([]sqlc.GetCacheMetricsDailyRow, error)
}

type SQLCRepository struct {
	queries *sqlc.Queries
}

func NewSQLCRepository(queries *sqlc.Queries) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (r *SQLCRepository) GetProjectMetrics(ctx context.Context, days string) ([]sqlc.GetProjectMetricsRow, error) {
	return r.queries.GetProjectMetrics(ctx, sql.NullString{String: days, Valid: true})
}

func (r *SQLCRepository) GetModelEfficiency(ctx context.Context, days string) ([]sqlc.GetModelEfficiencyRow, error) {
	return r.queries.GetModelEfficiency(ctx, sql.NullString{String: days, Valid: true})
}

func (r *SQLCRepository) GetCacheMetrics(ctx context.Context, days string) (sqlc.GetCacheMetricsRow, error) {
	return r.queries.GetCacheMetrics(ctx, sql.NullString{String: days, Valid: true})
}

func (r *SQLCRepository) GetCacheMetricsDaily(ctx context.Context, days string) ([]sqlc.GetCacheMetricsDailyRow, error) {
	return r.queries.GetCacheMetricsDaily(ctx, sql.NullString{String: days, Valid: true})
}
