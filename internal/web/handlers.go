package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/parser"
	"github.com/emiliopalmerini/mclaude/internal/util"
	"github.com/emiliopalmerini/mclaude/internal/web/templates"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	// Get stats
	startDate := time.Unix(0, 0).Format(time.RFC3339) // All time
	statsRow, _ := queries.GetAggregateStats(ctx, startDate)

	stats := templates.DashboardStats{
		SessionCount: statsRow.SessionCount,
		TotalTokens:  util.ToInt64(statsRow.TotalTokenInput) + util.ToInt64(statsRow.TotalTokenOutput),
		TotalCost:    util.ToFloat64(statsRow.TotalCostUsd),
		TotalTurns:   util.ToInt64(statsRow.TotalTurns),
		TokenInput:   util.ToInt64(statsRow.TotalTokenInput),
		TokenOutput:  util.ToInt64(statsRow.TotalTokenOutput),
		CacheRead:    util.ToInt64(statsRow.TotalTokenCacheRead),
		CacheWrite:   util.ToInt64(statsRow.TotalTokenCacheWrite),
		TotalErrors:  util.ToInt64(statsRow.TotalErrors),
	}

	// Get usage limit stats
	if planConfig, err := s.planConfigRepo.Get(ctx); err == nil && planConfig != nil {
		usageStats := &templates.UsageLimitStats{
			PlanType:    planConfig.PlanType,
			WindowHours: planConfig.WindowHours,
		}

		// Get the 5-hour token limit (learned or estimated)
		if planConfig.LearnedTokenLimit != nil {
			usageStats.TokenLimit = *planConfig.LearnedTokenLimit
			usageStats.IsLearned = true
		} else if preset, ok := domain.PlanPresets[planConfig.PlanType]; ok {
			usageStats.TokenLimit = preset.TokenEstimate
		}

		// Get 5-hour rolling window usage
		if summary, err := s.planConfigRepo.GetRollingWindowSummary(ctx, planConfig.WindowHours); err == nil {
			usageStats.TokensUsed = summary.TotalTokens

			// Calculate percentage
			if usageStats.TokenLimit > 0 {
				usageStats.UsagePercent = (summary.TotalTokens / usageStats.TokenLimit) * 100
			}

			// Determine status
			usageStats.Status = domain.GetStatusFromPercent(usageStats.UsagePercent)

			// Approximate minutes left (rolling window refreshes continuously)
			usageStats.MinutesLeft = planConfig.WindowHours * 60
		}

		// Get the weekly token limit (learned or estimated)
		if planConfig.WeeklyLearnedTokenLimit != nil {
			usageStats.WeeklyTokenLimit = *planConfig.WeeklyLearnedTokenLimit
			usageStats.WeeklyIsLearned = true
		} else if preset, ok := domain.WeeklyPlanPresets[planConfig.PlanType]; ok {
			usageStats.WeeklyTokenLimit = preset.TokenEstimate
		}

		// Get weekly rolling window usage
		if weeklySummary, err := s.planConfigRepo.GetWeeklyWindowSummary(ctx); err == nil {
			usageStats.WeeklyTokensUsed = weeklySummary.TotalTokens

			// Calculate percentage
			if usageStats.WeeklyTokenLimit > 0 {
				usageStats.WeeklyUsagePercent = (weeklySummary.TotalTokens / usageStats.WeeklyTokenLimit) * 100
			}

			// Determine status
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
		successCount := int64(0)
		failureCount := int64(0)
		if qualityStats.SuccessCount.Valid {
			successCount = int64(qualityStats.SuccessCount.Float64)
		}
		if qualityStats.FailureCount.Valid {
			failureCount = int64(qualityStats.FailureCount.Float64)
		}
		total := successCount + failureCount
		if total > 0 {
			rate := float64(successCount) / float64(total)
			stats.SuccessRate = &rate
		}
	}

	templates.Dashboard(stats).Render(ctx, w)
}

func (s *Server) handleSessions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	// Get sessions with metrics in single query
	sessions, _ := queries.ListSessionsWithMetrics(ctx, 50)

	// Build quality lookup map
	qualityMap := make(map[string]sqlc.ListSessionQualitiesForSessionsRow)
	if qualities, err := queries.ListSessionQualitiesForSessions(ctx); err == nil {
		for _, q := range qualities {
			qualityMap[q.SessionID] = q
		}
	}

	sessionList := make([]templates.SessionSummary, 0, len(sessions))
	for _, sess := range sessions {
		summary := templates.SessionSummary{
			ID:         sess.ID,
			CreatedAt:  sess.CreatedAt,
			ExitReason: sess.ExitReason,
			ProjectID:  sess.ProjectID,
			Turns:      sess.TurnCount,
			Tokens:     sess.TotalTokens,
		}
		if sess.ExperimentID.Valid {
			summary.ExperimentID = sess.ExperimentID.String
		}
		if sess.CostEstimateUsd.Valid {
			summary.Cost = sess.CostEstimateUsd.Float64
		}
		// Add quality data
		if q, ok := qualityMap[sess.ID]; ok {
			summary.IsReviewed = true
			if q.OverallRating.Valid {
				summary.OverallRating = int(q.OverallRating.Int64)
			}
			if q.IsSuccess.Valid {
				isSuccess := q.IsSuccess.Int64 == 1
				summary.IsSuccess = &isSuccess
			}
		}
		sessionList = append(sessionList, summary)
	}

	templates.Sessions(sessionList).Render(ctx, w)
}

func (s *Server) handleSessionDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	queries := sqlc.New(s.db)

	session, err := queries.GetSessionByID(ctx, id)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	detail := templates.SessionDetail{
		ID:             session.ID,
		ProjectID:      session.ProjectID,
		Cwd:            session.Cwd,
		PermissionMode: session.PermissionMode,
		ExitReason:     session.ExitReason,
		CreatedAt:      session.CreatedAt,
	}

	if session.ExperimentID.Valid {
		detail.ExperimentID = session.ExperimentID.String
	}
	if session.StartedAt.Valid {
		detail.StartedAt = session.StartedAt.String
	}
	if session.EndedAt.Valid {
		detail.EndedAt = session.EndedAt.String
	}
	if session.DurationSeconds.Valid {
		detail.DurationSeconds = session.DurationSeconds.Int64
	}

	// Get metrics
	if m, err := queries.GetSessionMetricsBySessionID(ctx, id); err == nil {
		detail.MessageCountUser = m.MessageCountUser
		detail.MessageCountAssistant = m.MessageCountAssistant
		detail.TurnCount = m.TurnCount
		detail.TokenInput = m.TokenInput
		detail.TokenOutput = m.TokenOutput
		detail.TokenCacheRead = m.TokenCacheRead
		detail.TokenCacheWrite = m.TokenCacheWrite
		detail.ErrorCount = m.ErrorCount
		if m.CostEstimateUsd.Valid {
			detail.CostEstimateUsd = m.CostEstimateUsd.Float64
		}
	}

	// Get tools
	tools, _ := queries.ListSessionToolsBySessionID(ctx, id)
	for _, t := range tools {
		detail.Tools = append(detail.Tools, templates.ToolUsage{
			Name:  t.ToolName,
			Count: t.InvocationCount,
		})
	}

	// Get files
	files, _ := queries.ListSessionFilesBySessionID(ctx, id)
	for _, f := range files {
		detail.Files = append(detail.Files, templates.FileOperation{
			Path:      f.FilePath,
			Operation: f.Operation,
			Count:     f.OperationCount,
		})
	}

	// Get quality
	if q, err := queries.GetSessionQualityBySessionID(ctx, id); err == nil {
		quality := templates.SessionQuality{
			SessionID: q.SessionID,
		}
		if q.OverallRating.Valid {
			quality.OverallRating = int(q.OverallRating.Int64)
		}
		if q.IsSuccess.Valid {
			isSuccess := q.IsSuccess.Int64 == 1
			quality.IsSuccess = &isSuccess
		}
		if q.AccuracyRating.Valid {
			quality.AccuracyRating = int(q.AccuracyRating.Int64)
		}
		if q.HelpfulnessRating.Valid {
			quality.HelpfulnessRating = int(q.HelpfulnessRating.Int64)
		}
		if q.EfficiencyRating.Valid {
			quality.EfficiencyRating = int(q.EfficiencyRating.Int64)
		}
		if q.Notes.Valid {
			quality.Notes = q.Notes.String
		}
		if q.ReviewedAt.Valid {
			quality.ReviewedAt = q.ReviewedAt.String
		}
		detail.Quality = &quality
	}

	templates.SessionDetailPage(detail).Render(ctx, w)
}

func (s *Server) handleExperiments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	// Get experiments with stats
	expStats, _ := queries.GetStatsForAllExperiments(ctx)
	statsMap := make(map[string]sqlc.GetStatsForAllExperimentsRow)
	for _, es := range expStats {
		statsMap[es.ExperimentID] = es
	}

	exps, _ := queries.ListExperiments(ctx)

	experiments := make([]templates.Experiment, 0, len(exps))
	for _, e := range exps {
		exp := templates.Experiment{
			ID:        e.ID,
			Name:      e.Name,
			IsActive:  e.IsActive == 1,
			StartedAt: e.StartedAt,
			CreatedAt: e.CreatedAt,
		}
		if e.Description.Valid {
			exp.Description = e.Description.String
		}
		if e.Hypothesis.Valid {
			exp.Hypothesis = e.Hypothesis.String
		}
		if e.EndedAt.Valid {
			exp.EndedAt = e.EndedAt.String
		}

		// Add stats
		if es, ok := statsMap[e.ID]; ok {
			exp.SessionCount = es.SessionCount
			exp.TotalTokens = util.ToInt64(es.TotalTokenInput) + util.ToInt64(es.TotalTokenOutput)
			exp.TotalCost = util.ToFloat64(es.TotalCostUsd)
			if es.SessionCount > 0 {
				exp.TokensPerSess = exp.TotalTokens / es.SessionCount
				exp.CostPerSession = exp.TotalCost / float64(es.SessionCount)
			}
		}

		experiments = append(experiments, exp)
	}

	templates.Experiments(experiments).Render(ctx, w)
}

func (s *Server) handleExperimentDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	queries := sqlc.New(s.db)

	// Get experiment
	exp, err := queries.GetExperimentByID(ctx, id)
	if err != nil {
		http.Error(w, "Experiment not found", http.StatusNotFound)
		return
	}

	detail := templates.ExperimentDetail{
		ID:        exp.ID,
		Name:      exp.Name,
		IsActive:  exp.IsActive == 1,
		StartedAt: exp.StartedAt,
		CreatedAt: exp.CreatedAt,
	}
	if exp.Description.Valid {
		detail.Description = exp.Description.String
	}
	if exp.Hypothesis.Valid {
		detail.Hypothesis = exp.Hypothesis.String
	}
	if exp.EndedAt.Valid {
		detail.EndedAt = exp.EndedAt.String
	}

	// Get aggregate stats
	statsRow, err := queries.GetAggregateStatsByExperiment(ctx, sqlc.GetAggregateStatsByExperimentParams{
		ExperimentID: util.NullString(exp.ID),
		CreatedAt:    "1970-01-01T00:00:00Z",
	})
	if err == nil {
		detail.SessionCount = statsRow.SessionCount
		detail.TotalTurns = util.ToInt64(statsRow.TotalTurns)
		detail.UserMessages = util.ToInt64(statsRow.TotalUserMessages)
		detail.AssistantMessages = util.ToInt64(statsRow.TotalAssistantMessages)
		detail.TotalErrors = util.ToInt64(statsRow.TotalErrors)
		detail.TokenInput = util.ToInt64(statsRow.TotalTokenInput)
		detail.TokenOutput = util.ToInt64(statsRow.TotalTokenOutput)
		detail.CacheRead = util.ToInt64(statsRow.TotalTokenCacheRead)
		detail.CacheWrite = util.ToInt64(statsRow.TotalTokenCacheWrite)
		detail.TotalTokens = detail.TokenInput + detail.TokenOutput
		detail.TotalCost = util.ToFloat64(statsRow.TotalCostUsd)
		if statsRow.SessionCount > 0 {
			detail.TokensPerSession = detail.TotalTokens / statsRow.SessionCount
			detail.CostPerSession = detail.TotalCost / float64(statsRow.SessionCount)
		}
	}

	// Get top tools for this experiment
	tools, _ := queries.GetTopToolsUsageByExperiment(ctx, sqlc.GetTopToolsUsageByExperimentParams{
		ExperimentID: util.NullString(exp.ID),
		Limit:        5,
	})
	for _, t := range tools {
		if t.TotalInvocations.Valid {
			detail.TopTools = append(detail.TopTools, templates.ToolUsage{
				Name:  t.ToolName,
				Count: int64(t.TotalInvocations.Float64),
			})
		}
	}

	// Get recent sessions for this experiment (with metrics in single query)
	sessions, _ := queries.ListSessionsWithMetricsByExperiment(ctx, sqlc.ListSessionsWithMetricsByExperimentParams{
		ExperimentID: util.NullString(exp.ID),
		Limit:        10,
	})
	for _, sess := range sessions {
		summary := templates.SessionSummary{
			ID:        sess.ID,
			CreatedAt: sess.CreatedAt,
			Turns:     sess.TurnCount,
			Tokens:    sess.TotalTokens,
		}
		if sess.CostEstimateUsd.Valid {
			summary.Cost = sess.CostEstimateUsd.Float64
		}
		detail.RecentSessions = append(detail.RecentSessions, summary)
	}

	// Get quality stats
	qualityStats, err := queries.GetQualityStatsByExperiment(ctx, util.NullString(exp.ID))
	if err == nil && qualityStats.ReviewedCount > 0 {
		detail.ReviewedCount = qualityStats.ReviewedCount
		if qualityStats.AvgOverallRating.Valid {
			avg := qualityStats.AvgOverallRating.Float64
			detail.AvgOverall = &avg
		}
		if qualityStats.AvgAccuracy.Valid {
			avg := qualityStats.AvgAccuracy.Float64
			detail.AvgAccuracy = &avg
		}
		if qualityStats.AvgHelpfulness.Valid {
			avg := qualityStats.AvgHelpfulness.Float64
			detail.AvgHelpfulness = &avg
		}
		if qualityStats.AvgEfficiency.Valid {
			avg := qualityStats.AvgEfficiency.Float64
			detail.AvgEfficiency = &avg
		}
		// Calculate success rate
		successCount := int64(0)
		failureCount := int64(0)
		if qualityStats.SuccessCount.Valid {
			successCount = int64(qualityStats.SuccessCount.Float64)
		}
		if qualityStats.FailureCount.Valid {
			failureCount = int64(qualityStats.FailureCount.Float64)
		}
		total := successCount + failureCount
		if total > 0 {
			rate := float64(successCount) / float64(total)
			detail.SuccessRate = &rate
		}
	}

	templates.ExperimentDetailPage(detail).Render(ctx, w)
}

func (s *Server) handleExperimentCompare(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	// Get experiment IDs from query params (e.g., ?ids=id1,id2,id3)
	idsParam := r.URL.Query().Get("ids")
	if idsParam == "" {
		// No experiments selected, show empty comparison
		templates.ExperimentComparePage(templates.ExperimentComparison{}).Render(ctx, w)
		return
	}

	ids := splitIDs(idsParam)
	if len(ids) < 2 {
		templates.ExperimentComparePage(templates.ExperimentComparison{}).Render(ctx, w)
		return
	}

	var items []templates.ExperimentCompareItem

	for _, id := range ids {
		exp, err := queries.GetExperimentByID(ctx, id)
		if err != nil {
			continue
		}

		item := templates.ExperimentCompareItem{
			Name:     exp.Name,
			IsActive: exp.IsActive == 1,
		}

		// Get aggregate stats
		statsRow, err := queries.GetAggregateStatsByExperiment(ctx, sqlc.GetAggregateStatsByExperimentParams{
			ExperimentID: util.NullString(exp.ID),
			CreatedAt:    "1970-01-01T00:00:00Z",
		})
		if err == nil {
			item.SessionCount = statsRow.SessionCount
			item.TotalTurns = util.ToInt64(statsRow.TotalTurns)
			item.UserMessages = util.ToInt64(statsRow.TotalUserMessages)
			item.AssistantMessages = util.ToInt64(statsRow.TotalAssistantMessages)
			item.TotalErrors = util.ToInt64(statsRow.TotalErrors)
			item.TokenInput = util.ToInt64(statsRow.TotalTokenInput)
			item.TokenOutput = util.ToInt64(statsRow.TotalTokenOutput)
			item.CacheRead = util.ToInt64(statsRow.TotalTokenCacheRead)
			item.CacheWrite = util.ToInt64(statsRow.TotalTokenCacheWrite)
			item.TotalTokens = item.TokenInput + item.TokenOutput
			item.TotalCost = util.ToFloat64(statsRow.TotalCostUsd)
			if statsRow.SessionCount > 0 {
				item.TokensPerSession = item.TotalTokens / statsRow.SessionCount
				item.CostPerSession = item.TotalCost / float64(statsRow.SessionCount)
			}
		}

		// Get quality stats
		qualityStats, err := queries.GetQualityStatsByExperiment(ctx, util.NullString(exp.ID))
		if err == nil && qualityStats.ReviewedCount > 0 {
			item.ReviewedCount = qualityStats.ReviewedCount

			if qualityStats.AvgOverallRating.Valid {
				avg := qualityStats.AvgOverallRating.Float64
				item.AvgOverall = &avg
			}
			if qualityStats.AvgAccuracy.Valid {
				avg := qualityStats.AvgAccuracy.Float64
				item.AvgAccuracy = &avg
			}
			if qualityStats.AvgHelpfulness.Valid {
				avg := qualityStats.AvgHelpfulness.Float64
				item.AvgHelpfulness = &avg
			}
			if qualityStats.AvgEfficiency.Valid {
				avg := qualityStats.AvgEfficiency.Float64
				item.AvgEfficiency = &avg
			}

			// Calculate success rate
			successCount := int64(0)
			failureCount := int64(0)
			if qualityStats.SuccessCount.Valid {
				successCount = int64(qualityStats.SuccessCount.Float64)
			}
			if qualityStats.FailureCount.Valid {
				failureCount = int64(qualityStats.FailureCount.Float64)
			}
			total := successCount + failureCount
			if total > 0 {
				rate := float64(successCount) / float64(total)
				item.SuccessRate = &rate
			}
		}

		items = append(items, item)
	}

	templates.ExperimentComparePage(templates.ExperimentComparison{
		Experiments: items,
	}).Render(ctx, w)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	pricing, _ := queries.ListModelPricing(ctx)

	models := make([]templates.ModelPricing, 0, len(pricing))
	for _, p := range pricing {
		model := templates.ModelPricing{
			ID:               p.ID,
			DisplayName:      p.DisplayName,
			InputPerMillion:  p.InputPerMillion,
			OutputPerMillion: p.OutputPerMillion,
			IsDefault:        p.IsDefault == 1,
		}
		if p.CacheReadPerMillion.Valid {
			model.CacheReadPerMillion = p.CacheReadPerMillion.Float64
		}
		if p.CacheWritePerMillion.Valid {
			model.CacheWritePerMillion = p.CacheWritePerMillion.Float64
		}
		models = append(models, model)
	}

	templates.Settings(models).Render(ctx, w)
}

// API Handlers

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

	stats := map[string]interface{}{
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
		switch d := stat.Date.(type) {
		case time.Time:
			labels[i] = d.Format("2006-01-02")
		case string:
			labels[i] = d
		default:
			labels[i] = fmt.Sprintf("%v", stat.Date)
		}
		tokens[i] = util.ToInt64(stat.TotalTokens)
		sessions[i] = stat.SessionCount
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
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
		switch d := stat.Date.(type) {
		case time.Time:
			labels[i] = d.Format("2006-01-02")
		case string:
			labels[i] = d
		default:
			labels[i] = fmt.Sprintf("%v", stat.Date)
		}
		costs[i] = util.ToFloat64(stat.TotalCost)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"labels": labels,
		"costs":  costs,
	})
}

func (s *Server) handleAPICreateExperiment(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (s *Server) handleAPIActivateExperiment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	queries := sqlc.New(s.db)

	queries.DeactivateAllExperiments(ctx)
	if err := queries.ActivateExperiment(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/experiments")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAPIDeactivateExperiment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	queries := sqlc.New(s.db)

	if err := queries.DeactivateExperiment(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/experiments")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAPIDeleteExperiment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	queries := sqlc.New(s.db)

	if err := queries.DeleteExperiment(ctx, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/experiments")
	w.WriteHeader(http.StatusOK)
}

// Session Review Handlers

func (s *Server) handleSessionReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	queries := sqlc.New(s.db)

	// Get session
	session, err := queries.GetSessionByID(ctx, id)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	// Build session detail
	detail := templates.SessionDetail{
		ID:             session.ID,
		ProjectID:      session.ProjectID,
		Cwd:            session.Cwd,
		PermissionMode: session.PermissionMode,
		ExitReason:     session.ExitReason,
		CreatedAt:      session.CreatedAt,
	}

	if session.ExperimentID.Valid {
		detail.ExperimentID = session.ExperimentID.String
	}
	if session.StartedAt.Valid {
		detail.StartedAt = session.StartedAt.String
	}
	if session.EndedAt.Valid {
		detail.EndedAt = session.EndedAt.String
	}
	if session.DurationSeconds.Valid {
		detail.DurationSeconds = session.DurationSeconds.Int64
	}

	// Get metrics
	if m, err := queries.GetSessionMetricsBySessionID(ctx, id); err == nil {
		detail.MessageCountUser = m.MessageCountUser
		detail.MessageCountAssistant = m.MessageCountAssistant
		detail.TurnCount = m.TurnCount
		detail.TokenInput = m.TokenInput
		detail.TokenOutput = m.TokenOutput
		detail.TokenCacheRead = m.TokenCacheRead
		detail.TokenCacheWrite = m.TokenCacheWrite
		detail.ErrorCount = m.ErrorCount
		if m.CostEstimateUsd.Valid {
			detail.CostEstimateUsd = m.CostEstimateUsd.Float64
		}
	}

	// Get existing quality review
	var quality templates.SessionQuality
	if q, err := s.qualityRepo.GetBySessionID(ctx, id); err == nil && q != nil {
		quality = convertDomainQualityToTemplate(q)
	}

	// Get transcript
	var transcriptMessages []templates.TranscriptMessage
	if s.transcriptStorage != nil {
		data, err := s.transcriptStorage.Get(ctx, id)
		if err == nil {
			messages, _ := parser.ParseTranscriptForViewer(data)
			transcriptMessages = convertViewerMessagesToTemplate(messages)
		}
	}

	viewData := templates.SessionReviewData{
		SessionDetail: detail,
		Quality:       quality,
		Transcript:    transcriptMessages,
	}

	templates.SessionReviewPage(viewData).Render(ctx, w)
}

func (s *Server) handleAPISaveQuality(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	quality := &domain.SessionQuality{
		SessionID: id,
	}

	// Parse overall rating
	if v := r.FormValue("overall_rating"); v != "" && v != "0" {
		if rating, err := strconv.Atoi(v); err == nil && rating >= 1 && rating <= 5 {
			quality.OverallRating = &rating
		}
	}

	// Parse is_success
	if v := r.FormValue("is_success"); v != "" {
		success := v == "1"
		quality.IsSuccess = &success
	}

	// Parse dimension ratings
	if v := r.FormValue("accuracy_rating"); v != "" && v != "0" {
		if rating, err := strconv.Atoi(v); err == nil && rating >= 1 && rating <= 5 {
			quality.AccuracyRating = &rating
		}
	}
	if v := r.FormValue("helpfulness_rating"); v != "" && v != "0" {
		if rating, err := strconv.Atoi(v); err == nil && rating >= 1 && rating <= 5 {
			quality.HelpfulnessRating = &rating
		}
	}
	if v := r.FormValue("efficiency_rating"); v != "" && v != "0" {
		if rating, err := strconv.Atoi(v); err == nil && rating >= 1 && rating <= 5 {
			quality.EfficiencyRating = &rating
		}
	}

	// Parse notes
	if v := r.FormValue("notes"); v != "" {
		quality.Notes = &v
	}

	// Set reviewed_at if any rating is provided
	if quality.OverallRating != nil || quality.IsSuccess != nil {
		now := time.Now()
		quality.ReviewedAt = &now
	}

	// Save to database
	if err := s.qualityRepo.Upsert(ctx, quality); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success indicator for HTMX
	templates.QualitySavedIndicator().Render(ctx, w)
}

func convertDomainQualityToTemplate(q *domain.SessionQuality) templates.SessionQuality {
	tq := templates.SessionQuality{
		SessionID: q.SessionID,
		IsSuccess: q.IsSuccess,
	}
	if q.OverallRating != nil {
		tq.OverallRating = *q.OverallRating
	}
	if q.AccuracyRating != nil {
		tq.AccuracyRating = *q.AccuracyRating
	}
	if q.HelpfulnessRating != nil {
		tq.HelpfulnessRating = *q.HelpfulnessRating
	}
	if q.EfficiencyRating != nil {
		tq.EfficiencyRating = *q.EfficiencyRating
	}
	if q.Notes != nil {
		tq.Notes = *q.Notes
	}
	if q.ReviewedAt != nil {
		tq.ReviewedAt = q.ReviewedAt.Format(time.RFC3339)
	}
	return tq
}

func convertViewerMessagesToTemplate(messages []parser.ViewerMessage) []templates.TranscriptMessage {
	result := make([]templates.TranscriptMessage, len(messages))
	for i, m := range messages {
		result[i] = templates.TranscriptMessage{
			Role:      m.Role,
			Content:   m.Content,
			Timestamp: m.Timestamp,
		}
		for _, t := range m.Tools {
			result[i].Tools = append(result[i].Tools, templates.TranscriptToolUse{
				Name:  t.Name,
				Input: t.Input,
			})
		}
	}
	return result
}

// Real-time API

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

// Helpers

func splitIDs(s string) []string {
	if s == "" {
		return nil
	}
	var ids []string
	for _, id := range strings.Split(s, ",") {
		id = strings.TrimSpace(id)
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}
