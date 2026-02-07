package web

import (
	"context"
	"database/sql"
	"testing"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/migrate"
	_ "github.com/tursodatabase/go-libsql"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("libsql", "file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	if err := migrate.RunAll(context.Background(), db); err != nil {
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
		repos.Pricing, repos.Sessions, repos.Metrics,
		repos.Stats, repos.Projects,
	)
}

func TestFetchDashboardData_EmptyDB(t *testing.T) {
	db := testDB(t)
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
