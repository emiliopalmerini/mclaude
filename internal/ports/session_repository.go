package ports

import (
	"context"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByID(ctx context.Context, id string) (*domain.Session, error)
	List(ctx context.Context, opts ListSessionsOptions) ([]*domain.Session, error)
	ListWithMetrics(ctx context.Context, opts ListSessionsOptions) ([]*domain.SessionListItem, error)
	Delete(ctx context.Context, id string) error
	DeleteBefore(ctx context.Context, before string) (int64, error)
	DeleteByProject(ctx context.Context, projectID string) (int64, error)
	DeleteByExperiment(ctx context.Context, experimentID string) (int64, error)
	GetTranscriptPathsBefore(ctx context.Context, before string) ([]domain.TranscriptPathInfo, error)
	GetTranscriptPathsByProject(ctx context.Context, projectID string) ([]domain.TranscriptPathInfo, error)
	GetTranscriptPathsByExperiment(ctx context.Context, experimentID string) ([]domain.TranscriptPathInfo, error)
}

type ListSessionsOptions struct {
	Limit        int
	ProjectID    *string
	ExperimentID *string
}

type SessionMetricsRepository interface {
	Create(ctx context.Context, metrics *domain.SessionMetrics) error
	GetBySessionID(ctx context.Context, sessionID string) (*domain.SessionMetrics, error)
}

type SessionToolRepository interface {
	CreateBatch(ctx context.Context, tools []*domain.SessionTool) error
	ListBySessionID(ctx context.Context, sessionID string) ([]*domain.SessionTool, error)
}

type SessionFileRepository interface {
	CreateBatch(ctx context.Context, files []*domain.SessionFile) error
	ListBySessionID(ctx context.Context, sessionID string) ([]*domain.SessionFile, error)
}

type SessionCommandRepository interface {
	CreateBatch(ctx context.Context, commands []*domain.SessionCommand) error
	ListBySessionID(ctx context.Context, sessionID string) ([]*domain.SessionCommand, error)
}

type SessionSubagentRepository interface {
	CreateBatch(ctx context.Context, subagents []*domain.SessionSubagent) error
	ListBySessionID(ctx context.Context, sessionID string) ([]*domain.SessionSubagent, error)
}

type ToolEventRepository interface {
	ListBySessionID(ctx context.Context, sessionID string) ([]*domain.ToolEvent, error)
}
