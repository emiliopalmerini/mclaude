package web

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/migrate"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	_ "github.com/tursodatabase/go-libsql"
)

func testTursoDB(t *testing.T) *sql.DB {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/tursodatabase/libsql-server:latest",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/health").WithPort("8080/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start Turso container: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	mappedPort, err := container.MappedPort(ctx, "8080")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}

	url := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())
	db, err := sql.Open("libsql", url)
	if err != nil {
		t.Fatalf("failed to connect to Turso: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping Turso: %v", err)
	}

	if err := migrate.RunAll(ctx, db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func testServer(t *testing.T, db *sql.DB) *Server {
	t.Helper()

	repos := turso.NewRepositories(db)
	return NewServer(
		db, 0, nil,
		repos.Quality, repos.PlanConfig, repos.Experiments,
		repos.ExperimentVariables, repos.Pricing, repos.Sessions, repos.Metrics,
		repos.Stats, repos.Projects,
	)
}

// seedTestData inserts experiments, projects, sessions, metrics, and tools.
// Returns experiment and project IDs.
func seedTestData(t *testing.T, db *sql.DB) (expID, projID string) {
	t.Helper()
	q := sqlc.New(db)
	ctx := context.Background()
	now := time.Now().UTC()

	expID = "exp-test-1"
	projID = "proj-test-1"

	// Create experiment
	_ = q.CreateExperiment(ctx, sqlc.CreateExperimentParams{
		ID:        expID,
		Name:      "Test Experiment",
		IsActive:  1,
		CreatedAt: now.Format(time.RFC3339),
	})

	// Create projects
	_, _ = db.ExecContext(ctx, `INSERT INTO projects (id, path, name, created_at) VALUES (?, ?, ?, ?)`,
		projID, "/home/test/proj", "test-proj", now.Format(time.RFC3339))
	_, _ = db.ExecContext(ctx, `INSERT INTO projects (id, path, name, created_at) VALUES (?, ?, ?, ?)`,
		"proj-test-2", "/home/test/other", "other-proj", now.Format(time.RFC3339))

	// Session 1: linked to experiment and project, created recently
	sess1Time := now.Add(-1 * time.Hour).Format(time.RFC3339)
	_ = q.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:             "sess-1",
		ProjectID:      projID,
		ExperimentID:   sql.NullString{String: expID, Valid: true},
		Cwd:            "/home/test/proj",
		PermissionMode: "default",
		ExitReason:     "exit",
		CreatedAt:      sess1Time,
		StartedAt:      sql.NullString{String: sess1Time, Valid: true},
	})
	_ = q.CreateSessionMetrics(ctx, sqlc.CreateSessionMetricsParams{
		SessionID:       "sess-1",
		TurnCount:       5,
		TokenInput:      1000,
		TokenOutput:     500,
		TokenCacheRead:  200,
		TokenCacheWrite: 100,
		CostEstimateUsd: sql.NullFloat64{Float64: 0.05, Valid: true},
	})
	_ = q.CreateSessionTool(ctx, sqlc.CreateSessionToolParams{
		SessionID:       "sess-1",
		ToolName:        "Read",
		InvocationCount: 10,
	})
	_ = q.CreateSessionTool(ctx, sqlc.CreateSessionToolParams{
		SessionID:       "sess-1",
		ToolName:        "Edit",
		InvocationCount: 3,
	})

	// Session 2: different project, no experiment
	sess2Time := now.Add(-30 * time.Minute).Format(time.RFC3339)
	_ = q.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:             "sess-2",
		ProjectID:      "proj-test-2",
		Cwd:            "/home/test/other",
		PermissionMode: "default",
		ExitReason:     "exit",
		CreatedAt:      sess2Time,
		StartedAt:      sql.NullString{String: sess2Time, Valid: true},
	})
	_ = q.CreateSessionMetrics(ctx, sqlc.CreateSessionMetricsParams{
		SessionID:       "sess-2",
		TurnCount:       3,
		TokenInput:      2000,
		TokenOutput:     800,
		TokenCacheRead:  400,
		TokenCacheWrite: 200,
		CostEstimateUsd: sql.NullFloat64{Float64: 0.10, Valid: true},
	})
	_ = q.CreateSessionTool(ctx, sqlc.CreateSessionToolParams{
		SessionID:       "sess-2",
		ToolName:        "Read",
		InvocationCount: 5,
	})

	return expID, projID
}

func TestFetchDashboardData_EmptyDB(t *testing.T) {
	db := testTursoDB(t)
	s := testServer(t, db)
	ctx := context.Background()

	stats := s.fetchDashboardData(ctx, dashboardFilters{})

	if stats.SessionCount != 0 {
		t.Errorf("expected 0 sessions, got %d", stats.SessionCount)
	}
	if stats.TotalTokens != 0 {
		t.Errorf("expected 0 tokens, got %d", stats.TotalTokens)
	}
	if stats.TotalCost != 0 {
		t.Errorf("expected 0 cost, got %f", stats.TotalCost)
	}
	if len(stats.TopTools) != 0 {
		t.Errorf("expected 0 top tools, got %d", len(stats.TopTools))
	}
	if len(stats.RecentSessions) != 0 {
		t.Errorf("expected 0 recent sessions, got %d", len(stats.RecentSessions))
	}
	if stats.ActiveExperiment != "" {
		t.Errorf("expected empty active experiment, got %s", stats.ActiveExperiment)
	}
	// Migration 016 seeds default model pricing
	if stats.DefaultModel != "Claude Sonnet 4" {
		t.Errorf("expected 'Claude Sonnet 4' (seeded by migration), got %q", stats.DefaultModel)
	}
}

func TestFetchDashboardData_NoFilter(t *testing.T) {
	db := testTursoDB(t)
	s := testServer(t, db)
	ctx := context.Background()

	seedTestData(t, db)

	stats := s.fetchDashboardData(ctx, dashboardFilters{})

	// Aggregate: both sessions (1000+500 + 2000+800 = 4300 total tokens)
	if stats.SessionCount != 2 {
		t.Errorf("expected 2 sessions, got %d", stats.SessionCount)
	}
	if stats.TotalTokens != 4300 {
		t.Errorf("expected 4300 total tokens, got %d", stats.TotalTokens)
	}
	if stats.TotalTurns != 8 {
		t.Errorf("expected 8 turns, got %d", stats.TotalTurns)
	}
	if stats.TotalCost < 0.149 || stats.TotalCost > 0.151 {
		t.Errorf("expected ~0.15 cost, got %f", stats.TotalCost)
	}

	// Filter dropdowns populated
	if len(stats.Experiments) != 1 {
		t.Errorf("expected 1 experiment in dropdown, got %d", len(stats.Experiments))
	}
	if len(stats.Projects) != 2 {
		t.Errorf("expected 2 projects in dropdown, got %d", len(stats.Projects))
	}

	// Top tools: Read=15, Edit=3
	if len(stats.TopTools) != 2 {
		t.Errorf("expected 2 top tools, got %d", len(stats.TopTools))
	}
	if len(stats.TopTools) > 0 && stats.TopTools[0].Name != "Read" {
		t.Errorf("expected top tool 'Read', got %q", stats.TopTools[0].Name)
	}
	if len(stats.TopTools) > 0 && stats.TopTools[0].Count != 15 {
		t.Errorf("expected Read count 15, got %d", stats.TopTools[0].Count)
	}

	// Recent sessions
	if len(stats.RecentSessions) != 2 {
		t.Errorf("expected 2 recent sessions, got %d", len(stats.RecentSessions))
	}

	// Active experiment
	if stats.ActiveExperiment != "Test Experiment" {
		t.Errorf("expected 'Test Experiment', got %q", stats.ActiveExperiment)
	}
}

func TestFetchDashboardData_ExperimentFilter(t *testing.T) {
	db := testTursoDB(t)
	s := testServer(t, db)
	ctx := context.Background()

	expID, _ := seedTestData(t, db)

	stats := s.fetchDashboardData(ctx, dashboardFilters{Experiment: expID})

	// Only session 1 is in the experiment
	if stats.SessionCount != 1 {
		t.Errorf("expected 1 session, got %d", stats.SessionCount)
	}
	if stats.TotalTokens != 1500 {
		t.Errorf("expected 1500 tokens, got %d", stats.TotalTokens)
	}
	if stats.FilterExperiment != expID {
		t.Errorf("expected filter experiment %q, got %q", expID, stats.FilterExperiment)
	}
}

func TestFetchDashboardData_ProjectFilter(t *testing.T) {
	db := testTursoDB(t)
	s := testServer(t, db)
	ctx := context.Background()

	_, projID := seedTestData(t, db)

	stats := s.fetchDashboardData(ctx, dashboardFilters{Project: projID})

	// Only session 1 is in proj-test-1
	if stats.SessionCount != 1 {
		t.Errorf("expected 1 session, got %d", stats.SessionCount)
	}
	if stats.TotalTokens != 1500 {
		t.Errorf("expected 1500 tokens, got %d", stats.TotalTokens)
	}
	if stats.FilterProject != projID {
		t.Errorf("expected filter project %q, got %q", projID, stats.FilterProject)
	}
}
