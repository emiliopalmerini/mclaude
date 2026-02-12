package web

import (
	"context"
	"net/http"

	"golang.org/x/sync/errgroup"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/util"
	"github.com/emiliopalmerini/mclaude/internal/web/templates"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
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
	queries := sqlc.New(s.db)
	startDate := util.GetStartDateForPeriod(filters.Period)

	// Results collected by each goroutine (no mutex needed — each writes to its own var)
	var (
		aggStats       *domain.AggregateStats
		experiments    []*domain.Experiment
		projects       []*domain.Project
		activeExp      sqlc.Experiment
		defaultModel   sqlc.ModelPricing
		tools          []sqlc.GetTopToolsUsageRow
		sessions       []sqlc.ListSessionsWithMetricsRow
		qualityStats   sqlc.GetOverallQualityStatsRow
		qualityStatsOK bool
	)

	g, gctx := errgroup.WithContext(ctx)

	// 1. Aggregate stats
	g.Go(func() error {
		var err error
		if filters.Experiment != "" {
			aggStats, err = s.statsRepo.GetAggregateByExperiment(gctx, filters.Experiment, startDate)
		} else if filters.Project != "" {
			aggStats, err = s.statsRepo.GetAggregateByProject(gctx, filters.Project, startDate)
		} else {
			aggStats, err = s.statsRepo.GetAggregate(gctx, startDate)
		}
		_ = err
		return nil
	})

	// 2. Experiments list (for dropdown)
	g.Go(func() error {
		experiments, _ = s.experimentRepo.List(gctx)
		return nil
	})

	// 3. Projects list (for dropdown)
	g.Go(func() error {
		projects, _ = s.projectRepo.List(gctx)
		return nil
	})

	// 4. Active experiment
	g.Go(func() error {
		activeExp, _ = queries.GetActiveExperiment(gctx)
		return nil
	})

	// 6. Default model
	g.Go(func() error {
		defaultModel, _ = queries.GetDefaultModelPricing(gctx)
		return nil
	})

	// 7. Top tools
	g.Go(func() error {
		tools, _ = queries.GetTopToolsUsage(gctx, sqlc.GetTopToolsUsageParams{
			CreatedAt: startDate,
			Limit:     5,
		})
		return nil
	})

	// 8. Recent sessions
	g.Go(func() error {
		sessions, _ = queries.ListSessionsWithMetrics(gctx, 5)
		return nil
	})

	// 9. Quality stats
	g.Go(func() error {
		var err error
		qualityStats, err = queries.GetOverallQualityStats(gctx)
		qualityStatsOK = err == nil
		return nil
	})

	_ = g.Wait()

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

	if activeExp.Name != "" {
		stats.ActiveExperiment = activeExp.Name
	}
	if defaultModel.DisplayName != "" {
		stats.DefaultModel = defaultModel.DisplayName
	}

	topTools := make([]templates.ToolUsage, 0, len(tools))
	for _, t := range tools {
		if t.TotalInvocations.Valid {
			topTools = append(topTools, templates.ToolUsage{
				Name:  t.ToolName,
				Count: int64(t.TotalInvocations.Float64),
			})
		}
	}
	stats.TopTools = topTools

	recentSessions := make([]templates.SessionSummary, 0, len(sessions))
	for _, sess := range sessions {
		summary := templates.SessionSummary{
			ID:         sess.ID,
			CreatedAt:  sess.CreatedAt,
			ExitReason: sess.ExitReason,
			Turns:      sess.TurnCount,
			Tokens:     sess.TotalTokens,
		}
		if sess.CostEstimateUsd.Valid {
			summary.Cost = sess.CostEstimateUsd.Float64
		}
		recentSessions = append(recentSessions, summary)
	}
	stats.RecentSessions = recentSessions

	if qualityStatsOK && qualityStats.ReviewedCount > 0 {
		stats.ReviewedCount = qualityStats.ReviewedCount
		if qualityStats.AvgOverallRating.Valid {
			avg := qualityStats.AvgOverallRating.Float64
			stats.AvgOverall = &avg
		}
		stats.SuccessRate = calculateSuccessRate(qualityStats.SuccessCount, qualityStats.FailureCount)
	}

	return stats
}
