package turso

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/emiliopalmerini/claude-watcher/internal/domain"
	"github.com/emiliopalmerini/claude-watcher/internal/ports"
	"github.com/emiliopalmerini/claude-watcher/sqlc/generated"
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
		ExperimentID:         nullString(session.ExperimentID),
		TranscriptPath:       session.TranscriptPath,
		TranscriptStoredPath: nullString(session.TranscriptStoredPath),
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
			ExperimentID: nullString(opts.ExperimentID),
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
	return r.queries.DeleteSessionsByExperiment(ctx, nullString(&experimentID))
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
		ExperimentID:         nullStringPtr(row.ExperimentID),
		TranscriptPath:       row.TranscriptPath,
		TranscriptStoredPath: nullStringPtr(row.TranscriptStoredPath),
		Cwd:                  row.Cwd,
		PermissionMode:       row.PermissionMode,
		ExitReason:           row.ExitReason,
		StartedAt:            startedAt,
		EndedAt:              endedAt,
		DurationSeconds:      durationSeconds,
		CreatedAt:            createdAt,
	}
}
