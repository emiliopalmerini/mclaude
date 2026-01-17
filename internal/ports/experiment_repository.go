package ports

import (
	"context"

	"github.com/emiliopalmerini/claude-watcher/internal/domain"
)

type ExperimentRepository interface {
	Create(ctx context.Context, experiment *domain.Experiment) error
	GetByID(ctx context.Context, id string) (*domain.Experiment, error)
	GetByName(ctx context.Context, name string) (*domain.Experiment, error)
	GetActive(ctx context.Context) (*domain.Experiment, error)
	List(ctx context.Context) ([]*domain.Experiment, error)
	Update(ctx context.Context, experiment *domain.Experiment) error
	Delete(ctx context.Context, id string) error
	Activate(ctx context.Context, id string) error
	Deactivate(ctx context.Context, id string) error
	DeactivateAll(ctx context.Context) error
}
