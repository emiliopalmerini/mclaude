package dashboard

import (
	"context"
	"database/sql"
	"encoding/json"

	"claude-watcher/internal/database/sqlc"
)

// Repository defines the data access interface for dashboard metrics.
type Repository interface {
	GetDashboardMetrics(ctx context.Context) (sqlc.GetDashboardMetricsRow, error)
	GetTodayMetrics(ctx context.Context) (sqlc.GetTodayMetricsRow, error)
	GetWeekMetrics(ctx context.Context) (sqlc.GetWeekMetricsRow, error)
	GetCacheMetrics(ctx context.Context, days string) (sqlc.GetCacheMetricsRow, error)
	GetTopProject(ctx context.Context) (sqlc.GetTopProjectRow, error)
	GetEfficiencyMetrics(ctx context.Context, days string) (sqlc.GetEfficiencyMetricsRow, error)
	GetToolsBreakdownAll(ctx context.Context, days string) ([]sql.NullString, error)
}

// SQLCRepository implements Repository using sqlc.Queries.
type SQLCRepository struct {
	queries *sqlc.Queries
}

// NewSQLCRepository creates a new SQLCRepository.
func NewSQLCRepository(queries *sqlc.Queries) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (r *SQLCRepository) GetDashboardMetrics(ctx context.Context) (sqlc.GetDashboardMetricsRow, error) {
	return r.queries.GetDashboardMetrics(ctx)
}

func (r *SQLCRepository) GetTodayMetrics(ctx context.Context) (sqlc.GetTodayMetricsRow, error) {
	return r.queries.GetTodayMetrics(ctx)
}

func (r *SQLCRepository) GetWeekMetrics(ctx context.Context) (sqlc.GetWeekMetricsRow, error) {
	return r.queries.GetWeekMetrics(ctx)
}

func (r *SQLCRepository) GetCacheMetrics(ctx context.Context, days string) (sqlc.GetCacheMetricsRow, error) {
	return r.queries.GetCacheMetrics(ctx, sql.NullString{String: days, Valid: true})
}

func (r *SQLCRepository) GetTopProject(ctx context.Context) (sqlc.GetTopProjectRow, error) {
	return r.queries.GetTopProject(ctx)
}

func (r *SQLCRepository) GetEfficiencyMetrics(ctx context.Context, days string) (sqlc.GetEfficiencyMetricsRow, error) {
	return r.queries.GetEfficiencyMetrics(ctx, sql.NullString{String: days, Valid: true})
}

func (r *SQLCRepository) GetToolsBreakdownAll(ctx context.Context, days string) ([]sql.NullString, error) {
	return r.queries.GetToolsBreakdownAll(ctx, sql.NullString{String: days, Valid: true})
}

// TopTool finds the most used tool from a list of tools breakdown rows.
func TopTool(rows []sql.NullString) string {
	totals := make(map[string]int)
	for _, row := range rows {
		if row.Valid && row.String != "" {
			tools := make(map[string]int)
			if err := json.Unmarshal([]byte(row.String), &tools); err == nil {
				for tool, count := range tools {
					totals[tool] += count
				}
			}
		}
	}
	if len(totals) == 0 {
		return "-"
	}
	maxTool := ""
	maxCount := 0
	for tool, count := range totals {
		if count > maxCount {
			maxCount = count
			maxTool = tool
		}
	}
	return maxTool
}
