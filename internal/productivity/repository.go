package productivity

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"

	"claude-watcher/internal/database/sqlc"
)

type Repository interface {
	GetEfficiencyMetrics(ctx context.Context, days string) (sqlc.GetEfficiencyMetricsRow, error)
	GetEfficiencyMetricsDaily(ctx context.Context, days string) ([]sqlc.GetEfficiencyMetricsDailyRow, error)
	GetDayOfWeekDistribution(ctx context.Context, days string) ([]sqlc.GetDayOfWeekDistributionRow, error)
	GetHourOfDayDistribution(ctx context.Context, hours string) ([]sqlc.GetHourOfDayDistributionRow, error)
	GetToolsBreakdownAll(ctx context.Context, days string) ([]sql.NullString, error)
}

type SQLCRepository struct {
	queries *sqlc.Queries
}

func NewSQLCRepository(queries *sqlc.Queries) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (r *SQLCRepository) GetEfficiencyMetrics(ctx context.Context, days string) (sqlc.GetEfficiencyMetricsRow, error) {
	return r.queries.GetEfficiencyMetrics(ctx, sql.NullString{String: days, Valid: true})
}

func (r *SQLCRepository) GetEfficiencyMetricsDaily(ctx context.Context, days string) ([]sqlc.GetEfficiencyMetricsDailyRow, error) {
	return r.queries.GetEfficiencyMetricsDaily(ctx, sql.NullString{String: days, Valid: true})
}

func (r *SQLCRepository) GetDayOfWeekDistribution(ctx context.Context, days string) ([]sqlc.GetDayOfWeekDistributionRow, error) {
	return r.queries.GetDayOfWeekDistribution(ctx, sql.NullString{String: days, Valid: true})
}

func (r *SQLCRepository) GetHourOfDayDistribution(ctx context.Context, hours string) ([]sqlc.GetHourOfDayDistributionRow, error) {
	return r.queries.GetHourOfDayDistribution(ctx, sql.NullString{String: hours, Valid: true})
}

func (r *SQLCRepository) GetToolsBreakdownAll(ctx context.Context, days string) ([]sql.NullString, error) {
	return r.queries.GetToolsBreakdownAll(ctx, sql.NullString{String: days, Valid: true})
}

// AggregateTools combines all tool usage from multiple sessions into a sorted list.
func AggregateTools(rows []sql.NullString, limit int) []ToolCount {
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

	result := make([]ToolCount, 0, len(totals))
	for name, count := range totals {
		result = append(result, ToolCount{Name: name, Count: count})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Count > result[j].Count
	})

	if limit > 0 && len(result) > limit {
		return result[:limit]
	}
	return result
}
