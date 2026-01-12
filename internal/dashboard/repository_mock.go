package dashboard

import (
	"context"
	"database/sql"

	"claude-watcher/internal/database/sqlc"
)

// MockRepository is a mock implementation of Repository for testing.
type MockRepository struct {
	GetDashboardMetricsFunc  func(ctx context.Context) (sqlc.GetDashboardMetricsRow, error)
	GetTodayMetricsFunc      func(ctx context.Context) (sqlc.GetTodayMetricsRow, error)
	GetWeekMetricsFunc       func(ctx context.Context) (sqlc.GetWeekMetricsRow, error)
	GetCacheMetricsFunc      func(ctx context.Context, days string) (sqlc.GetCacheMetricsRow, error)
	GetTopProjectFunc        func(ctx context.Context) (sqlc.GetTopProjectRow, error)
	GetEfficiencyMetricsFunc func(ctx context.Context, days string) (sqlc.GetEfficiencyMetricsRow, error)
	GetToolsBreakdownAllFunc func(ctx context.Context, days string) ([]sql.NullString, error)
}

func (m *MockRepository) GetDashboardMetrics(ctx context.Context) (sqlc.GetDashboardMetricsRow, error) {
	if m.GetDashboardMetricsFunc != nil {
		return m.GetDashboardMetricsFunc(ctx)
	}
	return sqlc.GetDashboardMetricsRow{}, nil
}

func (m *MockRepository) GetTodayMetrics(ctx context.Context) (sqlc.GetTodayMetricsRow, error) {
	if m.GetTodayMetricsFunc != nil {
		return m.GetTodayMetricsFunc(ctx)
	}
	return sqlc.GetTodayMetricsRow{}, nil
}

func (m *MockRepository) GetWeekMetrics(ctx context.Context) (sqlc.GetWeekMetricsRow, error) {
	if m.GetWeekMetricsFunc != nil {
		return m.GetWeekMetricsFunc(ctx)
	}
	return sqlc.GetWeekMetricsRow{}, nil
}

func (m *MockRepository) GetCacheMetrics(ctx context.Context, days string) (sqlc.GetCacheMetricsRow, error) {
	if m.GetCacheMetricsFunc != nil {
		return m.GetCacheMetricsFunc(ctx, days)
	}
	return sqlc.GetCacheMetricsRow{}, nil
}

func (m *MockRepository) GetTopProject(ctx context.Context) (sqlc.GetTopProjectRow, error) {
	if m.GetTopProjectFunc != nil {
		return m.GetTopProjectFunc(ctx)
	}
	return sqlc.GetTopProjectRow{}, nil
}

func (m *MockRepository) GetEfficiencyMetrics(ctx context.Context, days string) (sqlc.GetEfficiencyMetricsRow, error) {
	if m.GetEfficiencyMetricsFunc != nil {
		return m.GetEfficiencyMetricsFunc(ctx, days)
	}
	return sqlc.GetEfficiencyMetricsRow{}, nil
}

func (m *MockRepository) GetToolsBreakdownAll(ctx context.Context, days string) ([]sql.NullString, error) {
	if m.GetToolsBreakdownAllFunc != nil {
		return m.GetToolsBreakdownAllFunc(ctx, days)
	}
	return nil, nil
}
