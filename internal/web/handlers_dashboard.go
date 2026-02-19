package web

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/ports"
	"github.com/emiliopalmerini/mclaude/internal/util"
	"github.com/emiliopalmerini/mclaude/internal/web/templates"
)

type dashboardFilters struct {
	Period     string
	Experiment string
	Project    string
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filters := dashboardFilters{
		Period:     r.URL.Query().Get("period"),
		Experiment: r.URL.Query().Get("experiment"),
		Project:    r.URL.Query().Get("project"),
	}
	stats := s.fetchDashboardData(ctx, filters)
	_ = templates.Dashboard(stats).Render(ctx, w)
}

func (s *Server) fetchDashboardData(ctx context.Context, filters dashboardFilters) templates.DashboardStats {
	startDate := util.GetStartDateForPeriod(filters.Period)

	// Sequential fetching — the remote libsql driver doesn't support concurrent connections well,
	// and SetMaxOpenConns(1) serializes access anyway.

	// 1. Aggregate stats
	var aggStats *domain.AggregateStats
	var err error
	if filters.Experiment != "" {
		aggStats, err = s.statsRepo.GetAggregateByExperiment(ctx, filters.Experiment, startDate)
	} else if filters.Project != "" {
		aggStats, err = s.statsRepo.GetAggregateByProject(ctx, filters.Project, startDate)
	} else {
		aggStats, err = s.statsRepo.GetAggregate(ctx, startDate)
	}
	if err != nil {
		slog.Error("dashboard: aggregate stats", "error", err)
	}

	// 2. Experiments list (for dropdown)
	experiments, err := s.experimentRepo.List(ctx)
	if err != nil {
		slog.Error("dashboard: experiments list", "error", err)
	}

	// 3. Projects list (for dropdown)
	projects, err := s.projectRepo.List(ctx)
	if err != nil {
		slog.Error("dashboard: projects list", "error", err)
	}

	// 4. Active experiment (via port interface)
	activeExp, err := s.experimentRepo.GetActive(ctx)
	if err != nil {
		slog.Error("dashboard: active experiment", "error", err)
	}

	// 5. Default model (via port interface)
	defaultModel, err := s.pricingRepo.GetDefault(ctx)
	if err != nil {
		slog.Error("dashboard: default model", "error", err)
	}

	// 6. Top tools (via port interface)
	tools, err := s.statsRepo.GetTopTools(ctx, startDate, 5)
	if err != nil {
		slog.Error("dashboard: top tools", "error", err)
	}

	// 7. Recent sessions (via port interface)
	sessionItems, err := s.sessionRepo.ListWithMetrics(ctx, ports.ListSessionsOptions{Limit: 5})
	if err != nil {
		slog.Error("dashboard: recent sessions", "error", err)
	}

	// Assemble results
	stats := templates.DashboardStats{
		FilterPeriod:     filters.Period,
		FilterExperiment: filters.Experiment,
		FilterProject:    filters.Project,
	}

	if aggStats != nil {
		stats.SessionCount = aggStats.SessionCount
		stats.TotalTokens = aggStats.TotalTokenInput + aggStats.TotalTokenOutput
		stats.TotalCost = aggStats.TotalCostUsd
		stats.TotalTurns = aggStats.TotalTurns
		stats.TokenInput = aggStats.TotalTokenInput
		stats.TokenOutput = aggStats.TotalTokenOutput
		stats.CacheRead = aggStats.TotalTokenCacheRead
		stats.CacheWrite = aggStats.TotalTokenCacheWrite
		stats.TotalErrors = aggStats.TotalErrors
	}

	for _, e := range experiments {
		stats.Experiments = append(stats.Experiments, templates.FilterOption{ID: e.ID, Name: e.Name})
	}
	for _, p := range projects {
		stats.Projects = append(stats.Projects, templates.FilterOption{ID: p.ID, Name: p.Name})
	}

	if activeExp != nil {
		stats.ActiveExperiment = activeExp.Name
	}
	if defaultModel != nil {
		stats.DefaultModel = defaultModel.DisplayName
	}

	topTools := make([]templates.ToolUsage, 0, len(tools))
	for _, t := range tools {
		topTools = append(topTools, templates.ToolUsage{
			Name:  t.ToolName,
			Count: t.TotalInvocations,
		})
	}
	stats.TopTools = topTools

	recentSessions := make([]templates.SessionSummary, 0, len(sessionItems))
	for _, sess := range sessionItems {
		summary := templates.SessionSummary{
			ID:         sess.ID,
			CreatedAt:  sess.CreatedAt,
			ExitReason: sess.ExitReason,
			Turns:      sess.TurnCount,
			Tokens:     sess.TotalTokens,
		}
		if sess.Cost != nil {
			summary.Cost = *sess.Cost
		}
		recentSessions = append(recentSessions, summary)
	}
	stats.RecentSessions = recentSessions

	return stats
}
