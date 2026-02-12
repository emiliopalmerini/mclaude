package turso_test

import (
	"context"
	"testing"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

func TestExperimentVariableRepository(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	// Seed an experiment
	queries := sqlc.New(db)
	err := queries.CreateExperiment(ctx, sqlc.CreateExperimentParams{
		ID:        "exp-1",
		Name:      "test-experiment",
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		IsActive:  1,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("failed to seed experiment: %v", err)
	}

	repo := turso.NewExperimentVariableRepository(db)

	// List should be empty initially
	vars, err := repo.ListByExperimentID(ctx, "exp-1")
	if err != nil {
		t.Fatalf("ListByExperimentID failed: %v", err)
	}
	if len(vars) != 0 {
		t.Fatalf("expected 0 variables, got %d", len(vars))
	}

	// Set two variables
	if err := repo.Set(ctx, "exp-1", "model", "opus-4.6"); err != nil {
		t.Fatalf("Set model failed: %v", err)
	}
	if err := repo.Set(ctx, "exp-1", "permission_mode", "plan"); err != nil {
		t.Fatalf("Set permission_mode failed: %v", err)
	}

	// List should return 2, ordered by key
	vars, err = repo.ListByExperimentID(ctx, "exp-1")
	if err != nil {
		t.Fatalf("ListByExperimentID failed: %v", err)
	}
	if len(vars) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(vars))
	}
	if vars[0].Key != "model" || vars[0].Value != "opus-4.6" {
		t.Errorf("expected model=opus-4.6, got %s=%s", vars[0].Key, vars[0].Value)
	}
	if vars[1].Key != "permission_mode" || vars[1].Value != "plan" {
		t.Errorf("expected permission_mode=plan, got %s=%s", vars[1].Key, vars[1].Value)
	}

	// Upsert should update existing
	if err := repo.Set(ctx, "exp-1", "model", "sonnet-4.5"); err != nil {
		t.Fatalf("Set (upsert) failed: %v", err)
	}
	vars, err = repo.ListByExperimentID(ctx, "exp-1")
	if err != nil {
		t.Fatalf("ListByExperimentID failed: %v", err)
	}
	if len(vars) != 2 {
		t.Fatalf("expected 2 variables after upsert, got %d", len(vars))
	}
	if vars[0].Key != "model" || vars[0].Value != "sonnet-4.5" {
		t.Errorf("expected model=sonnet-4.5 after upsert, got %s=%s", vars[0].Key, vars[0].Value)
	}

	// Delete one variable
	if err := repo.Delete(ctx, "exp-1", "model"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	vars, err = repo.ListByExperimentID(ctx, "exp-1")
	if err != nil {
		t.Fatalf("ListByExperimentID failed: %v", err)
	}
	if len(vars) != 1 {
		t.Fatalf("expected 1 variable after delete, got %d", len(vars))
	}
	if vars[0].Key != "permission_mode" {
		t.Errorf("expected remaining variable to be permission_mode, got %s", vars[0].Key)
	}
}

func TestExperimentVariableRepository_CascadeDelete(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()

	queries := sqlc.New(db)
	err := queries.CreateExperiment(ctx, sqlc.CreateExperimentParams{
		ID:        "exp-cascade",
		Name:      "cascade-test",
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		IsActive:  1,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("failed to seed experiment: %v", err)
	}

	repo := turso.NewExperimentVariableRepository(db)
	if err := repo.Set(ctx, "exp-cascade", "key1", "val1"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Delete the experiment; variables should cascade
	if err := queries.DeleteExperiment(ctx, "exp-cascade"); err != nil {
		t.Fatalf("DeleteExperiment failed: %v", err)
	}

	vars, err := repo.ListByExperimentID(ctx, "exp-cascade")
	if err != nil {
		t.Fatalf("ListByExperimentID failed: %v", err)
	}
	if len(vars) != 0 {
		t.Fatalf("expected 0 variables after cascade delete, got %d", len(vars))
	}
}
