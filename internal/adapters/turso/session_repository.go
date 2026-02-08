package turso

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/ports"
	"github.com/emiliopalmerini/mclaude/internal/util"
	"github.com/emiliopalmerini/mclaude/sqlc/generated"
)

type SessionRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	var startedAt, endedAt sql.NullString
	if session.StartedAt != nil {
		startedAt = sql.NullString{String: session.StartedAt.Format(time.RFC3339), Valid: true}
	}
	if session.EndedAt != nil {
		endedAt = sql.NullString{String: session.EndedAt.Format(time.RFC3339), Valid: true}
	}

	var durationSeconds sql.NullInt64
	if session.DurationSeconds != nil {
		durationSeconds = sql.NullInt64{Int64: *session.DurationSeconds, Valid: true}
	}

	return r.queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:                   session.ID,
		ProjectID:            session.ProjectID,
		ExperimentID:         util.NullStringPtr(session.ExperimentID),
		TranscriptPath:       session.TranscriptPath,
		TranscriptStoredPath: util.NullStringPtr(session.TranscriptStoredPath),
		Cwd:                  session.Cwd,
		PermissionMode:       session.PermissionMode,
		ExitReason:           session.ExitReason,
		StartedAt:            startedAt,
		EndedAt:              endedAt,
		DurationSeconds:      durationSeconds,
		CreatedAt:            session.CreatedAt.Format(time.RFC3339),
	})
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (*domain.Session, error) {
	row, err := r.queries.GetSessionByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return sessionFromRow(row), nil
}

func (r *SessionRepository) List(ctx context.Context, opts ports.ListSessionsOptions) ([]*domain.Session, error) {
	limit := int64(opts.Limit)
	if limit == 0 {
		limit = 50
	}

	var rows []sqlc.Session
	var err error

	if opts.ProjectID != nil {
		rows, err = r.queries.ListSessionsByProject(ctx, sqlc.ListSessionsByProjectParams{
			ProjectID: *opts.ProjectID,
			Limit:     limit,
		})
	} else if opts.ExperimentID != nil {
		rows, err = r.queries.ListSessionsByExperiment(ctx, sqlc.ListSessionsByExperimentParams{
			ExperimentID: util.NullStringPtr(opts.ExperimentID),
			Limit:        limit,
		})
	} else {
		rows, err = r.queries.ListSessions(ctx, limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	sessions := make([]*domain.Session, len(rows))
	for i, row := range rows {
		sessions[i] = sessionFromRow(row)
	}
	return sessions, nil
}

func (r *SessionRepository) ListWithMetrics(ctx context.Context, opts ports.ListSessionsOptions) ([]*domain.SessionListItem, error) {
	limit := int64(opts.Limit)
	if limit == 0 {
		limit = 50
	}

	if opts.ProjectID != nil {
		rows, err := r.queries.ListSessionsWithMetricsFullByProject(ctx, sqlc.ListSessionsWithMetricsFullByProjectParams{
			ProjectID: *opts.ProjectID,
			Limit:     limit,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list sessions with metrics: %w", err)
		}
		items := make([]*domain.SessionListItem, len(rows))
		for i, row := range rows {
			items[i] = &domain.SessionListItem{
				ID:            row.ID,
				ProjectID:     row.ProjectID,
				ExperimentID:  util.NullStringToPtr(row.ExperimentID),
				ExitReason:    row.ExitReason,
				CreatedAt:     row.CreatedAt,
				Duration:      nullInt64ToPtr(row.DurationSeconds),
				TurnCount:     row.TurnCount,
				TotalTokens:   row.TotalTokens,
				Cost:          nullFloat64ToPtr(row.CostEstimateUsd),
				ModelID:       util.NullStringToPtr(row.ModelID),
				SubagentCount: row.SubagentCount,
			}
		}
		return items, nil
	}

	if opts.ExperimentID != nil {
		rows, err := r.queries.ListSessionsWithMetricsFullByExperiment(ctx, sqlc.ListSessionsWithMetricsFullByExperimentParams{
			ExperimentID: util.NullStringPtr(opts.ExperimentID),
			Limit:        limit,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list sessions with metrics: %w", err)
		}
		items := make([]*domain.SessionListItem, len(rows))
		for i, row := range rows {
			items[i] = &domain.SessionListItem{
				ID:            row.ID,
				ProjectID:     row.ProjectID,
				ExperimentID:  util.NullStringToPtr(row.ExperimentID),
				ExitReason:    row.ExitReason,
				CreatedAt:     row.CreatedAt,
				Duration:      nullInt64ToPtr(row.DurationSeconds),
				TurnCount:     row.TurnCount,
				TotalTokens:   row.TotalTokens,
				Cost:          nullFloat64ToPtr(row.CostEstimateUsd),
				ModelID:       util.NullStringToPtr(row.ModelID),
				SubagentCount: row.SubagentCount,
			}
		}
		return items, nil
	}

	rows, err := r.queries.ListSessionsWithMetricsFull(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions with metrics: %w", err)
	}
	items := make([]*domain.SessionListItem, len(rows))
	for i, row := range rows {
		items[i] = &domain.SessionListItem{
			ID:            row.ID,
			ProjectID:     row.ProjectID,
			ExperimentID:  util.NullStringToPtr(row.ExperimentID),
			ExitReason:    row.ExitReason,
			CreatedAt:     row.CreatedAt,
			Duration:      nullInt64ToPtr(row.DurationSeconds),
			TurnCount:     row.TurnCount,
			TotalTokens:   row.TotalTokens,
			Cost:          nullFloat64ToPtr(row.CostEstimateUsd),
			ModelID:       util.NullStringToPtr(row.ModelID),
			SubagentCount: row.SubagentCount,
		}
	}
	return items, nil
}

func nullInt64ToPtr(n sql.NullInt64) *int64 {
	if !n.Valid {
		return nil
	}
	return &n.Int64
}

func nullFloat64ToPtr(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	return &n.Float64
}

func (r *SessionRepository) Delete(ctx context.Context, id string) error {
	return r.queries.DeleteSession(ctx, id)
}

func (r *SessionRepository) DeleteBefore(ctx context.Context, before string) (int64, error) {
	return r.queries.DeleteSessionsBefore(ctx, before)
}

func (r *SessionRepository) DeleteByProject(ctx context.Context, projectID string) (int64, error) {
	return r.queries.DeleteSessionsByProject(ctx, projectID)
}

func (r *SessionRepository) DeleteByExperiment(ctx context.Context, experimentID string) (int64, error) {
	return r.queries.DeleteSessionsByExperiment(ctx, util.NullStringPtr(&experimentID))
}

func (r *SessionRepository) GetTranscriptPathsBefore(ctx context.Context, before string) ([]domain.TranscriptPathInfo, error) {
	rows, err := r.queries.GetSessionTranscriptPathsBefore(ctx, before)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript paths: %w", err)
	}
	paths := make([]domain.TranscriptPathInfo, len(rows))
	for i, row := range rows {
		paths[i] = domain.TranscriptPathInfo{ID: row.ID, TranscriptPath: row.TranscriptStoredPath.String}
	}
	return paths, nil
}

func (r *SessionRepository) GetTranscriptPathsByProject(ctx context.Context, projectID string) ([]domain.TranscriptPathInfo, error) {
	rows, err := r.queries.GetSessionTranscriptPathsByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript paths: %w", err)
	}
	paths := make([]domain.TranscriptPathInfo, len(rows))
	for i, row := range rows {
		paths[i] = domain.TranscriptPathInfo{ID: row.ID, TranscriptPath: row.TranscriptStoredPath.String}
	}
	return paths, nil
}

func (r *SessionRepository) GetTranscriptPathsByExperiment(ctx context.Context, experimentID string) ([]domain.TranscriptPathInfo, error) {
	rows, err := r.queries.GetSessionTranscriptPathsByExperiment(ctx, util.NullStringPtr(&experimentID))
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript paths: %w", err)
	}
	paths := make([]domain.TranscriptPathInfo, len(rows))
	for i, row := range rows {
		paths[i] = domain.TranscriptPathInfo{ID: row.ID, TranscriptPath: row.TranscriptStoredPath.String}
	}
	return paths, nil
}

func sessionFromRow(row sqlc.Session) *domain.Session {
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)

	var startedAt, endedAt *time.Time
	if row.StartedAt.Valid {
		t, _ := time.Parse(time.RFC3339, row.StartedAt.String)
		startedAt = &t
	}
	if row.EndedAt.Valid {
		t, _ := time.Parse(time.RFC3339, row.EndedAt.String)
		endedAt = &t
	}

	var durationSeconds *int64
	if row.DurationSeconds.Valid {
		durationSeconds = &row.DurationSeconds.Int64
	}

	return &domain.Session{
		ID:                   row.ID,
		ProjectID:            row.ProjectID,
		ExperimentID:         util.NullStringToPtr(row.ExperimentID),
		TranscriptPath:       row.TranscriptPath,
		TranscriptStoredPath: util.NullStringToPtr(row.TranscriptStoredPath),
		Cwd:                  row.Cwd,
		PermissionMode:       row.PermissionMode,
		ExitReason:           row.ExitReason,
		StartedAt:            startedAt,
		EndedAt:              endedAt,
		DurationSeconds:      durationSeconds,
		CreatedAt:            createdAt,
	}
}
