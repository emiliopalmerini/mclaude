package turso

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

type ExperimentVariableRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewExperimentVariableRepository(db *sql.DB) *ExperimentVariableRepository {
	return &ExperimentVariableRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *ExperimentVariableRepository) Set(ctx context.Context, experimentID, key, value string) error {
	return r.queries.UpsertExperimentVariable(ctx, sqlc.UpsertExperimentVariableParams{
		ExperimentID: experimentID,
		Key:          key,
		Value:        value,
	})
}

func (r *ExperimentVariableRepository) ListByExperimentID(ctx context.Context, experimentID string) ([]*domain.ExperimentVariable, error) {
	rows, err := r.queries.ListExperimentVariablesByExperimentID(ctx, experimentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list experiment variables: %w", err)
	}

	vars := make([]*domain.ExperimentVariable, len(rows))
	for i, row := range rows {
		vars[i] = &domain.ExperimentVariable{
			ID:           row.ID,
			ExperimentID: row.ExperimentID,
			Key:          row.Key,
			Value:        row.Value,
		}
	}
	return vars, nil
}

func (r *ExperimentVariableRepository) Delete(ctx context.Context, experimentID, key string) error {
	return r.queries.DeleteExperimentVariable(ctx, sqlc.DeleteExperimentVariableParams{
		ExperimentID: experimentID,
		Key:          key,
	})
}
