package web

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/util"
	"github.com/emiliopalmerini/mclaude/internal/web/templates"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

func (s *Server) handleAPIStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	period := r.URL.Query().Get("period")
	startDate := util.GetStartDateForPeriod(period)

	statsRow, err := queries.GetAggregateStats(ctx, startDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stats := map[string]any{
		"session_count": statsRow.SessionCount,
		"total_tokens":  util.ToInt64(statsRow.TotalTokenInput) + util.ToInt64(statsRow.TotalTokenOutput),
		"total_cost":    util.ToFloat64(statsRow.TotalCostUsd),
		"token_input":   util.ToInt64(statsRow.TotalTokenInput),
		"token_output":  util.ToInt64(statsRow.TotalTokenOutput),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleAPIChartTokens(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	period := r.URL.Query().Get("period")
	startDate := util.GetStartDateForPeriod(period)
	if period == "" {
		// Default to last 30 days for charts
		startDate = time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
	}

	stats, _ := queries.GetDailyStats(ctx, sqlc.GetDailyStatsParams{
		CreatedAt: startDate,
		Limit:     30,
	})

	labels := make([]string, len(stats))
	tokens := make([]int64, len(stats))
	sessions := make([]int64, len(stats))

	for i, stat := range stats {
		labels[i] = formatChartDate(stat.Date)
		tokens[i] = util.ToInt64(stat.TotalTokens)
		sessions[i] = stat.SessionCount
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"labels":   labels,
		"tokens":   tokens,
		"sessions": sessions,
	})
}

func (s *Server) handleAPIChartCost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	period := r.URL.Query().Get("period")
	startDate := util.GetStartDateForPeriod(period)
	if period == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
	}

	stats, _ := queries.GetDailyStats(ctx, sqlc.GetDailyStatsParams{
		CreatedAt: startDate,
		Limit:     30,
	})

	labels := make([]string, len(stats))
	costs := make([]float64, len(stats))

	for i, stat := range stats {
		labels[i] = formatChartDate(stat.Date)
		costs[i] = util.ToFloat64(stat.TotalCost)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"labels": labels,
		"costs":  costs,
	})
}

func (s *Server) handleAPIChartHeatmap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	// Get daily stats for the current year
	startDate := time.Now().AddDate(0, 0, -365).Format(time.RFC3339)
	stats, _ := queries.GetDailyStats(ctx, sqlc.GetDailyStatsParams{
		CreatedAt: startDate,
		Limit:     366,
	})

	data := make([][2]any, len(stats))
	for i, stat := range stats {
		data[i] = [2]any{formatChartDate(stat.Date), stat.SessionCount}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"data": data,
	})
}

func (s *Server) handleAPIRealtimeUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get plan config for limits
	planConfig, err := s.planConfigRepo.Get(ctx)
	if err != nil || planConfig == nil {
		// No plan configured
		if r.Header.Get("HX-Request") == "true" {
			// HTMX request - return HTML
			templates.UsageLimitContent(nil).Render(ctx, w)
			return
		}
		// JSON API request
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(templates.RealtimeUsageStats{Available: false})
		return
	}

	// Build UsageLimitStats for template rendering
	usageStats := &templates.UsageLimitStats{
		PlanType:    planConfig.PlanType,
		WindowHours: planConfig.WindowHours,
	}

	// Get 5-hour limit
	if planConfig.LearnedTokenLimit != nil {
		usageStats.TokenLimit = *planConfig.LearnedTokenLimit
		usageStats.IsLearned = true
	} else if preset, ok := domain.PlanPresets[planConfig.PlanType]; ok {
		usageStats.TokenLimit = preset.TokenEstimate
	}

	// Get weekly limit
	if planConfig.WeeklyLearnedTokenLimit != nil {
		usageStats.WeeklyTokenLimit = *planConfig.WeeklyLearnedTokenLimit
		usageStats.WeeklyIsLearned = true
	} else if preset, ok := domain.WeeklyPlanPresets[planConfig.PlanType]; ok {
		usageStats.WeeklyTokenLimit = preset.TokenEstimate
	}

	// Try Prometheus first for real-time data
	promAvailable := false
	if s.promClient.IsAvailable(ctx) {
		// 5-hour window
		if usage, err := s.promClient.GetRollingWindowUsage(ctx, planConfig.WindowHours); err == nil && usage.Available {
			usageStats.TokensUsed = usage.TotalTokens
			promAvailable = true
		}

		// 7-day window (168 hours)
		if usage, err := s.promClient.GetRollingWindowUsage(ctx, 168); err == nil && usage.Available {
			usageStats.WeeklyTokensUsed = usage.TotalTokens
		}
	}

	// Fall back to local DB if Prometheus not available
	if !promAvailable {
		if summary, err := s.planConfigRepo.GetRollingWindowSummary(ctx, planConfig.WindowHours); err == nil {
			usageStats.TokensUsed = summary.TotalTokens
		}
		if summary, err := s.planConfigRepo.GetWeeklyWindowSummary(ctx); err == nil {
			usageStats.WeeklyTokensUsed = summary.TotalTokens
		}
	}

	// Calculate percentages and status
	if usageStats.TokenLimit > 0 {
		usageStats.UsagePercent = (usageStats.TokensUsed / usageStats.TokenLimit) * 100
		usageStats.Status = domain.GetStatusFromPercent(usageStats.UsagePercent)
	}

	if usageStats.WeeklyTokenLimit > 0 {
		usageStats.WeeklyUsagePercent = (usageStats.WeeklyTokensUsed / usageStats.WeeklyTokenLimit) * 100
		usageStats.WeeklyStatus = domain.GetStatusFromPercent(usageStats.WeeklyUsagePercent)
	}

	usageStats.MinutesLeft = planConfig.WindowHours * 60

	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// HTMX request - return HTML
		templates.UsageLimitContent(usageStats).Render(ctx, w)
		return
	}

	// JSON API request - convert to RealtimeUsageStats
	realtimeStats := templates.RealtimeUsageStats{
		Available:       true,
		Source:          "local",
		FiveHourTokens:  usageStats.TokensUsed,
		WeeklyTokens:    usageStats.WeeklyTokensUsed,
		FiveHourPercent: usageStats.UsagePercent,
		WeeklyPercent:   usageStats.WeeklyUsagePercent,
		FiveHourStatus:  usageStats.Status,
		WeeklyStatus:    usageStats.WeeklyStatus,
		FiveHourLimit:   usageStats.TokenLimit,
		WeeklyLimit:     usageStats.WeeklyTokenLimit,
	}
	if promAvailable {
		realtimeStats.Source = "prometheus"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(realtimeStats)
}
