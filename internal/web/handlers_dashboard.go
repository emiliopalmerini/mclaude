package web

import (
	"context"
	"net/http"

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
	templates.Dashboard(stats).Render(ctx, w)
}

func (s *Server) fetchDashboardData(ctx context.Context, filters dashboardFilters) templates.DashboardStats {
	queries := sqlc.New(s.db)
	startDate := util.GetStartDateForPeriod(filters.Period)

	// Get aggregate stats based on filters
	var aggStats *domain.AggregateStats
	if filters.Experiment != "" {
		aggStats, _ = s.statsRepo.GetAggregateByExperiment(ctx, filters.Experiment, startDate)
	} else if filters.Project != "" {
		aggStats, _ = s.statsRepo.GetAggregateByProject(ctx, filters.Project, startDate)
	} else {
		aggStats, _ = s.statsRepo.GetAggregate(ctx, startDate)
	}

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

	// Populate filter dropdowns
	if experiments, err := s.experimentRepo.List(ctx); err == nil {
		for _, e := range experiments {
			stats.Experiments = append(stats.Experiments, templates.FilterOption{ID: e.ID, Name: e.Name})
		}
	}
	if projects, err := s.projectRepo.List(ctx); err == nil {
		for _, p := range projects {
			stats.Projects = append(stats.Projects, templates.FilterOption{ID: p.ID, Name: p.Name})
		}
	}

	// Get usage limit stats
	if planConfig, err := s.planConfigRepo.Get(ctx); err == nil && planConfig != nil {
		usageStats := &templates.UsageLimitStats{
			PlanType:    planConfig.PlanType,
			WindowHours: planConfig.WindowHours,
		}

		if planConfig.LearnedTokenLimit != nil {
			usageStats.TokenLimit = *planConfig.LearnedTokenLimit
			usageStats.IsLearned = true
		} else if preset, ok := domain.PlanPresets[planConfig.PlanType]; ok {
			usageStats.TokenLimit = preset.TokenEstimate
		}

		if summary, err := s.planConfigRepo.GetRollingWindowSummary(ctx, planConfig.WindowHours); err == nil {
			usageStats.TokensUsed = summary.TotalTokens
			if usageStats.TokenLimit > 0 {
				usageStats.UsagePercent = (summary.TotalTokens / usageStats.TokenLimit) * 100
			}
			usageStats.Status = domain.GetStatusFromPercent(usageStats.UsagePercent)
			usageStats.MinutesLeft = planConfig.WindowHours * 60
		}

		if planConfig.WeeklyLearnedTokenLimit != nil {
			usageStats.WeeklyTokenLimit = *planConfig.WeeklyLearnedTokenLimit
			usageStats.WeeklyIsLearned = true
		} else if preset, ok := domain.WeeklyPlanPresets[planConfig.PlanType]; ok {
			usageStats.WeeklyTokenLimit = preset.TokenEstimate
		}

		if weeklySummary, err := s.planConfigRepo.GetWeeklyWindowSummary(ctx); err == nil {
			usageStats.WeeklyTokensUsed = weeklySummary.TotalTokens
			if usageStats.WeeklyTokenLimit > 0 {
				usageStats.WeeklyUsagePercent = (weeklySummary.TotalTokens / usageStats.WeeklyTokenLimit) * 100
			}
			usageStats.WeeklyStatus = domain.GetStatusFromPercent(usageStats.WeeklyUsagePercent)
		}

		stats.UsageStats = usageStats
	}

	// Get active experiment
	activeExp, _ := queries.GetActiveExperiment(ctx)
	if activeExp.Name != "" {
		stats.ActiveExperiment = activeExp.Name
	}

	// Get default model
	defaultModel, _ := queries.GetDefaultModelPricing(ctx)
	if defaultModel.DisplayName != "" {
		stats.DefaultModel = defaultModel.DisplayName
	}

	// Get top tools
	tools, _ := queries.GetTopToolsUsage(ctx, sqlc.GetTopToolsUsageParams{
		CreatedAt: startDate,
		Limit:     5,
	})

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

	// Get recent sessions (with metrics in single query)
	sessions, _ := queries.ListSessionsWithMetrics(ctx, 5)
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

	// Get overall quality stats
	qualityStats, err := queries.GetOverallQualityStats(ctx)
	if err == nil && qualityStats.ReviewedCount > 0 {
		stats.ReviewedCount = qualityStats.ReviewedCount
		if qualityStats.AvgOverallRating.Valid {
			avg := qualityStats.AvgOverallRating.Float64
			stats.AvgOverall = &avg
		}
		stats.SuccessRate = calculateSuccessRate(qualityStats.SuccessCount, qualityStats.FailureCount)
	}

	return stats
}
