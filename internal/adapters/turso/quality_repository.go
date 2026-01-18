package turso

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/emiliopalmerini/claude-watcher/internal/domain"
	sqlc "github.com/emiliopalmerini/claude-watcher/sqlc/generated"
)

type SessionQualityRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewSessionQualityRepository(db *sql.DB) *SessionQualityRepository {
	return &SessionQualityRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *SessionQualityRepository) Upsert(ctx context.Context, q *domain.SessionQuality) error {
	params := sqlc.UpsertSessionQualityParams{
		SessionID: q.SessionID,
	}

	if q.OverallRating != nil {
		params.OverallRating = sql.NullInt64{Int64: int64(*q.OverallRating), Valid: true}
	}
	if q.IsSuccess != nil {
		val := int64(0)
		if *q.IsSuccess {
			val = 1
		}
		params.IsSuccess = sql.NullInt64{Int64: val, Valid: true}
	}
	if q.AccuracyRating != nil {
		params.AccuracyRating = sql.NullInt64{Int64: int64(*q.AccuracyRating), Valid: true}
	}
	if q.HelpfulnessRating != nil {
		params.HelpfulnessRating = sql.NullInt64{Int64: int64(*q.HelpfulnessRating), Valid: true}
	}
	if q.EfficiencyRating != nil {
		params.EfficiencyRating = sql.NullInt64{Int64: int64(*q.EfficiencyRating), Valid: true}
	}
	if q.Notes != nil {
		params.Notes = sql.NullString{String: *q.Notes, Valid: true}
	}
	if q.ReviewedAt != nil {
		params.ReviewedAt = sql.NullString{String: q.ReviewedAt.Format(time.RFC3339), Valid: true}
	}

	return r.queries.UpsertSessionQuality(ctx, params)
}

func (r *SessionQualityRepository) GetBySessionID(ctx context.Context, sessionID string) (*domain.SessionQuality, error) {
	row, err := r.queries.GetSessionQualityBySessionID(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get session quality: %w", err)
	}

	quality := &domain.SessionQuality{
		SessionID: row.SessionID,
		CreatedAt: parseTime(row.CreatedAt),
	}

	if row.OverallRating.Valid {
		val := int(row.OverallRating.Int64)
		quality.OverallRating = &val
	}
	if row.IsSuccess.Valid {
		val := row.IsSuccess.Int64 == 1
		quality.IsSuccess = &val
	}
	if row.AccuracyRating.Valid {
		val := int(row.AccuracyRating.Int64)
		quality.AccuracyRating = &val
	}
	if row.HelpfulnessRating.Valid {
		val := int(row.HelpfulnessRating.Int64)
		quality.HelpfulnessRating = &val
	}
	if row.EfficiencyRating.Valid {
		val := int(row.EfficiencyRating.Int64)
		quality.EfficiencyRating = &val
	}
	if row.Notes.Valid {
		quality.Notes = &row.Notes.String
	}
	if row.ReviewedAt.Valid {
		t, _ := time.Parse(time.RFC3339, row.ReviewedAt.String)
		quality.ReviewedAt = &t
	}

	return quality, nil
}

func (r *SessionQualityRepository) Delete(ctx context.Context, sessionID string) error {
	return r.queries.DeleteSessionQuality(ctx, sessionID)
}

func (r *SessionQualityRepository) ListUnreviewed(ctx context.Context, limit int) ([]string, error) {
	return r.queries.ListUnreviewedSessionIDs(ctx, int64(limit))
}

func parseTime(s string) time.Time {
	t, _ := time.Parse(time.RFC3339, s)
	return t
}
