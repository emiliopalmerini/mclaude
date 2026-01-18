package ports

import (
	"context"

	"github.com/emiliopalmerini/claude-watcher/internal/domain"
)

type SessionQualityRepository interface {
	Upsert(ctx context.Context, quality *domain.SessionQuality) error
	GetBySessionID(ctx context.Context, sessionID string) (*domain.SessionQuality, error)
	Delete(ctx context.Context, sessionID string) error
	ListUnreviewed(ctx context.Context, limit int) ([]string, error)
}
