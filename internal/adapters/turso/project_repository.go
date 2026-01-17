package turso

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"

	"github.com/emiliopalmerini/claude-watcher/internal/domain"
	"github.com/emiliopalmerini/claude-watcher/sqlc/generated"
)

type ProjectRepository struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	return r.queries.CreateProject(ctx, sqlc.CreateProjectParams{
		ID:        project.ID,
		Path:      project.Path,
		Name:      project.Name,
		CreatedAt: project.CreatedAt.Format(time.RFC3339),
	})
}

func (r *ProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	row, err := r.queries.GetProjectByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	return projectFromRow(row), nil
}

func (r *ProjectRepository) GetOrCreate(ctx context.Context, path string) (*domain.Project, error) {
	id := hashPath(path)

	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	project := &domain.Project{
		ID:        id,
		Path:      path,
		Name:      filepath.Base(path),
		CreatedAt: time.Now().UTC(),
	}

	if err := r.Create(ctx, project); err != nil {
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	return project, nil
}

func (r *ProjectRepository) List(ctx context.Context) ([]*domain.Project, error) {
	rows, err := r.queries.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	projects := make([]*domain.Project, len(rows))
	for i, row := range rows {
		projects[i] = projectFromRow(row)
	}
	return projects, nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	return r.queries.DeleteProject(ctx, id)
}

func projectFromRow(row sqlc.Project) *domain.Project {
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)
	return &domain.Project{
		ID:        row.ID,
		Path:      row.Path,
		Name:      row.Name,
		CreatedAt: createdAt,
	}
}

func hashPath(path string) string {
	hash := sha256.Sum256([]byte(path))
	return hex.EncodeToString(hash[:])
}
