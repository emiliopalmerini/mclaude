package sessions

import (
	"context"
	"database/sql"

	"claude-watcher/internal/database/sqlc"
)

// MockRepository is a mock implementation of Repository for testing.
type MockRepository struct {
	ListSessionsFunc          func(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error)
	CountSessionsFunc         func(ctx context.Context) (int64, error)
	ListSessionsFilteredFunc  func(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error)
	CountSessionsFilteredFunc func(ctx context.Context, params sqlc.CountSessionsFilteredParams) (int64, error)
	GetDistinctHostnamesFunc  func(ctx context.Context) ([]string, error)
	GetDistinctBranchesFunc   func(ctx context.Context) ([]sql.NullString, error)
	GetDistinctModelsFunc     func(ctx context.Context) ([]sql.NullString, error)
}

func (m *MockRepository) ListSessions(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error) {
	if m.ListSessionsFunc != nil {
		return m.ListSessionsFunc(ctx, params)
	}
	return []sqlc.ListSessionsRow{}, nil
}

func (m *MockRepository) CountSessions(ctx context.Context) (int64, error) {
	if m.CountSessionsFunc != nil {
		return m.CountSessionsFunc(ctx)
	}
	return 0, nil
}

func (m *MockRepository) ListSessionsFiltered(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error) {
	if m.ListSessionsFilteredFunc != nil {
		return m.ListSessionsFilteredFunc(ctx, params)
	}
	return []sqlc.ListSessionsFilteredRow{}, nil
}

func (m *MockRepository) CountSessionsFiltered(ctx context.Context, params sqlc.CountSessionsFilteredParams) (int64, error) {
	if m.CountSessionsFilteredFunc != nil {
		return m.CountSessionsFilteredFunc(ctx, params)
	}
	return 0, nil
}

func (m *MockRepository) GetDistinctHostnames(ctx context.Context) ([]string, error) {
	if m.GetDistinctHostnamesFunc != nil {
		return m.GetDistinctHostnamesFunc(ctx)
	}
	return []string{}, nil
}

func (m *MockRepository) GetDistinctBranches(ctx context.Context) ([]sql.NullString, error) {
	if m.GetDistinctBranchesFunc != nil {
		return m.GetDistinctBranchesFunc(ctx)
	}
	return []sql.NullString{}, nil
}

func (m *MockRepository) GetDistinctModels(ctx context.Context) ([]sql.NullString, error) {
	if m.GetDistinctModelsFunc != nil {
		return m.GetDistinctModelsFunc(ctx)
	}
	return []sql.NullString{}, nil
}
