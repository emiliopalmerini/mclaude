package cli

import (
	"context"
	"fmt"
	"testing"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

// mockExperimentRepo is a minimal mock for testing getExperimentByName.
type mockExperimentRepo struct {
	exp *domain.Experiment
	err error
}

func (m *mockExperimentRepo) Create(_ context.Context, _ *domain.Experiment) error { return nil }
func (m *mockExperimentRepo) GetByID(_ context.Context, _ string) (*domain.Experiment, error) {
	return nil, nil
}
func (m *mockExperimentRepo) GetByName(_ context.Context, _ string) (*domain.Experiment, error) {
	return m.exp, m.err
}
func (m *mockExperimentRepo) GetActive(_ context.Context) (*domain.Experiment, error) {
	return nil, nil
}
func (m *mockExperimentRepo) List(_ context.Context) ([]*domain.Experiment, error) { return nil, nil }
func (m *mockExperimentRepo) Update(_ context.Context, _ *domain.Experiment) error { return nil }
func (m *mockExperimentRepo) Delete(_ context.Context, _ string) error             { return nil }
func (m *mockExperimentRepo) Activate(_ context.Context, _ string) error           { return nil }
func (m *mockExperimentRepo) Deactivate(_ context.Context, _ string) error         { return nil }
func (m *mockExperimentRepo) DeactivateAll(_ context.Context) error                { return nil }

func TestGetExperimentByName_Found(t *testing.T) {
	repo := &mockExperimentRepo{exp: &domain.Experiment{ID: "abc", Name: "test"}}
	exp, err := getExperimentByName(context.Background(), repo, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exp.ID != "abc" {
		t.Errorf("expected ID abc, got %s", exp.ID)
	}
}

func TestGetExperimentByName_NotFound(t *testing.T) {
	repo := &mockExperimentRepo{exp: nil, err: nil}
	_, err := getExperimentByName(context.Background(), repo, "missing")
	if err == nil {
		t.Fatal("expected error for missing experiment")
	}
	if err.Error() != `experiment "missing" not found` {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestGetExperimentByName_RepoError(t *testing.T) {
	repo := &mockExperimentRepo{exp: nil, err: fmt.Errorf("db error")}
	_, err := getExperimentByName(context.Background(), repo, "test")
	if err == nil {
		t.Fatal("expected error for repo failure")
	}
}
