package ports

import (
	"context"

	"github.com/emiliopalmerini/claude-watcher/internal/domain"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	GetOrCreate(ctx context.Context, path string) (*domain.Project, error)
	List(ctx context.Context) ([]*domain.Project, error)
	Delete(ctx context.Context, id string) error
}
