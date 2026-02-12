package turso

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/util"
	"github.com/emiliopalmerini/mclaude/sqlc/generated"
)

type ExperimentRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewExperimentRepository(db *sql.DB) *ExperimentRepository {
	return &ExperimentRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *ExperimentRepository) Create(ctx context.Context, experiment *domain.Experiment) error {
	var endedAt sql.NullString
	if experiment.EndedAt != nil {
		endedAt = sql.NullString{String: experiment.EndedAt.Format(time.RFC3339), Valid: true}
	}

	return r.queries.CreateExperiment(ctx, sqlc.CreateExperimentParams{
		ID:          experiment.ID,
		Name:        experiment.Name,
		Description: util.NullStringPtr(experiment.Description),
		Hypothesis:  util.NullStringPtr(experiment.Hypothesis),
		StartedAt:   experiment.StartedAt.Format(time.RFC3339),
		EndedAt:     endedAt,
		IsActive:    util.BoolToInt64(experiment.IsActive),
		CreatedAt:   experiment.CreatedAt.Format(time.RFC3339),
		ModelID:     util.NullStringPtr(experiment.ModelID),
		PlanType:    util.NullStringPtr(experiment.PlanType),
		Notes:       util.NullStringPtr(experiment.Notes),
	})
}

func (r *ExperimentRepository) GetByID(ctx context.Context, id string) (*domain.Experiment, error) {
	row, err := r.queries.GetExperimentByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}
	return experimentFromRow(row), nil
}

func (r *ExperimentRepository) GetByName(ctx context.Context, name string) (*domain.Experiment, error) {
	row, err := r.queries.GetExperimentByName(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get experiment by name: %w", err)
	}
	return experimentFromRow(row), nil
}

func (r *ExperimentRepository) GetActive(ctx context.Context) (*domain.Experiment, error) {
	row, err := r.queries.GetActiveExperiment(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active experiment: %w", err)
	}
	return experimentFromRow(row), nil
}

func (r *ExperimentRepository) List(ctx context.Context) ([]*domain.Experiment, error) {
	rows, err := r.queries.ListExperiments(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list experiments: %w", err)
	}

	experiments := make([]*domain.Experiment, len(rows))
	for i, row := range rows {
		experiments[i] = experimentFromRow(row)
	}
	return experiments, nil
}

func (r *ExperimentRepository) Update(ctx context.Context, experiment *domain.Experiment) error {
	var endedAt sql.NullString
	if experiment.EndedAt != nil {
		endedAt = sql.NullString{String: experiment.EndedAt.Format(time.RFC3339), Valid: true}
	}

	return r.queries.UpdateExperiment(ctx, sqlc.UpdateExperimentParams{
		Name:        experiment.Name,
		Description: util.NullStringPtr(experiment.Description),
		Hypothesis:  util.NullStringPtr(experiment.Hypothesis),
		StartedAt:   experiment.StartedAt.Format(time.RFC3339),
		EndedAt:     endedAt,
		IsActive:    util.BoolToInt64(experiment.IsActive),
		ModelID:     util.NullStringPtr(experiment.ModelID),
		PlanType:    util.NullStringPtr(experiment.PlanType),
		Notes:       util.NullStringPtr(experiment.Notes),
		ID:          experiment.ID,
	})
}

func (r *ExperimentRepository) Delete(ctx context.Context, id string) error {
	return r.queries.DeleteExperiment(ctx, id)
}

func (r *ExperimentRepository) Activate(ctx context.Context, id string) error {
	return r.queries.ActivateExperiment(ctx, id)
}

func (r *ExperimentRepository) Deactivate(ctx context.Context, id string) error {
	return r.queries.DeactivateExperiment(ctx, id)
}

func (r *ExperimentRepository) DeactivateAll(ctx context.Context) error {
	return r.queries.DeactivateAllExperiments(ctx)
}

func experimentFromRow(row sqlc.Experiment) *domain.Experiment {
	startedAt, _ := time.Parse(time.RFC3339, row.StartedAt)
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)

	var endedAt *time.Time
	if row.EndedAt.Valid {
		t, _ := time.Parse(time.RFC3339, row.EndedAt.String)
		endedAt = &t
	}

	return &domain.Experiment{
		ID:          row.ID,
		Name:        row.Name,
		Description: util.NullStringToPtr(row.Description),
		Hypothesis:  util.NullStringToPtr(row.Hypothesis),
		StartedAt:   startedAt,
		EndedAt:     endedAt,
		IsActive:    row.IsActive == 1,
		CreatedAt:   createdAt,
		ModelID:     util.NullStringToPtr(row.ModelID),
		PlanType:    util.NullStringToPtr(row.PlanType),
		Notes:       util.NullStringToPtr(row.Notes),
	}
}
