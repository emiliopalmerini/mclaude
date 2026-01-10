package sessions

import (
	"context"
	"database/sql"

	"claude-watcher/internal/database/sqlc"
)

// Repository defines the data access interface for sessions list.
type Repository interface {
	ListSessions(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error)
	CountSessions(ctx context.Context) (int64, error)
	ListSessionsFiltered(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error)
	CountSessionsFiltered(ctx context.Context, params sqlc.CountSessionsFilteredParams) (int64, error)
	GetDistinctHostnames(ctx context.Context) ([]string, error)
	GetDistinctBranches(ctx context.Context) ([]sql.NullString, error)
	GetDistinctModels(ctx context.Context) ([]sql.NullString, error)
}

// SQLCRepository implements Repository using sqlc.Queries.
type SQLCRepository struct {
	queries *sqlc.Queries
}

// NewSQLCRepository creates a new SQLCRepository.
func NewSQLCRepository(queries *sqlc.Queries) *SQLCRepository {
	return &SQLCRepository{queries: queries}
}

func (r *SQLCRepository) ListSessions(ctx context.Context, params sqlc.ListSessionsParams) ([]sqlc.ListSessionsRow, error) {
	return r.queries.ListSessions(ctx, params)
}

func (r *SQLCRepository) CountSessions(ctx context.Context) (int64, error) {
	return r.queries.CountSessions(ctx)
}

func (r *SQLCRepository) ListSessionsFiltered(ctx context.Context, params sqlc.ListSessionsFilteredParams) ([]sqlc.ListSessionsFilteredRow, error) {
	return r.queries.ListSessionsFiltered(ctx, params)
}

func (r *SQLCRepository) CountSessionsFiltered(ctx context.Context, params sqlc.CountSessionsFilteredParams) (int64, error) {
	return r.queries.CountSessionsFiltered(ctx, params)
}

func (r *SQLCRepository) GetDistinctHostnames(ctx context.Context) ([]string, error) {
	return r.queries.GetDistinctHostnames(ctx)
}

func (r *SQLCRepository) GetDistinctBranches(ctx context.Context) ([]sql.NullString, error) {
	return r.queries.GetDistinctBranches(ctx)
}

func (r *SQLCRepository) GetDistinctModels(ctx context.Context) ([]sql.NullString, error) {
	return r.queries.GetDistinctModels(ctx)
}
