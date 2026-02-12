package ports

import (
	"context"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

type ExperimentVariableRepository interface {
	Set(ctx context.Context, experimentID, key, value string) error
	ListByExperimentID(ctx context.Context, experimentID string) ([]*domain.ExperimentVariable, error)
	Delete(ctx context.Context, experimentID, key string) error
}
