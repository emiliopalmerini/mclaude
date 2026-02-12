package web

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/util"
	"github.com/emiliopalmerini/mclaude/internal/web/templates"
	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

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
		if e.ModelID.Valid {
			exp.ModelID = e.ModelID.String
		}
		if e.PlanType.Valid {
			exp.PlanType = e.PlanType.String
		}
		if e.Notes.Valid {
			exp.Notes = e.Notes.String
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
	if exp.ModelID.Valid {
		detail.ModelID = exp.ModelID.String
	}
	if exp.PlanType.Valid {
		detail.PlanType = exp.PlanType.String
	}
	if exp.Notes.Valid {
		detail.Notes = exp.Notes.String
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

	// Compute normalized behavior metrics
	toolCalls, _ := s.statsRepo.GetTotalToolCallsByExperiment(ctx, exp.ID)
	agStats := &domain.AggregateStats{
		TotalTurns:           detail.TotalTurns,
		TotalTokenInput:      detail.TokenInput,
		TotalTokenOutput:     detail.TokenOutput,
		TotalTokenCacheRead:  detail.CacheRead,
		TotalTokenCacheWrite: detail.CacheWrite,
		TotalErrors:          detail.TotalErrors,
	}
	normalized := agStats.ComputeNormalized(toolCalls)
	detail.TokensPerTurn = normalized.TokensPerTurn
	detail.OutputRatio = normalized.OutputRatio
	detail.CacheHitRate = normalized.CacheHitRate
	detail.ErrorRate = normalized.ErrorRate
	detail.ToolCallsPerTurn = normalized.ToolCallsPerTurn

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
		detail.SuccessRate = calculateSuccessRate(qualityStats.SuccessCount, qualityStats.FailureCount)
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
		if exp.ModelID.Valid {
			item.ModelID = exp.ModelID.String
		}
		if exp.PlanType.Valid {
			item.PlanType = exp.PlanType.String
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

		// Compute normalized behavior metrics
		toolCalls, _ := s.statsRepo.GetTotalToolCallsByExperiment(ctx, exp.ID)
		agStats := &domain.AggregateStats{
			TotalTurns:           item.TotalTurns,
			TotalTokenInput:      item.TokenInput,
			TotalTokenOutput:     item.TokenOutput,
			TotalTokenCacheRead:  item.CacheRead,
			TotalTokenCacheWrite: item.CacheWrite,
			TotalErrors:          item.TotalErrors,
		}
		normalized := agStats.ComputeNormalized(toolCalls)
		item.TokensPerTurn = normalized.TokensPerTurn
		item.OutputRatio = normalized.OutputRatio
		item.CacheHitRate = normalized.CacheHitRate
		item.ErrorRate = normalized.ErrorRate
		item.ToolCallsPerTurn = normalized.ToolCallsPerTurn

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

			item.SuccessRate = calculateSuccessRate(qualityStats.SuccessCount, qualityStats.FailureCount)
		}

		items = append(items, item)
	}

	templates.ExperimentComparePage(templates.ExperimentComparison{
		Experiments: items,
	}).Render(ctx, w)
}

func (s *Server) handleAPICreateExperiment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	existing, err := s.experimentRepo.GetByName(ctx, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if existing != nil {
		http.Error(w, fmt.Sprintf("Experiment %q already exists", name), http.StatusConflict)
		return
	}

	if err := s.experimentRepo.DeactivateAll(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now().UTC()
	exp := &domain.Experiment{
		ID:        uuid.New().String(),
		Name:      name,
		StartedAt: now,
		IsActive:  true,
		CreatedAt: now,
	}
	if desc := strings.TrimSpace(r.FormValue("description")); desc != "" {
		exp.Description = &desc
	}
	if hyp := strings.TrimSpace(r.FormValue("hypothesis")); hyp != "" {
		exp.Hypothesis = &hyp
	}
	if model := strings.TrimSpace(r.FormValue("model_id")); model != "" {
		exp.ModelID = &model
	}
	if plan := strings.TrimSpace(r.FormValue("plan_type")); plan != "" {
		exp.PlanType = &plan
	}
	if notes := strings.TrimSpace(r.FormValue("notes")); notes != "" {
		exp.Notes = &notes
	}

	if err := s.experimentRepo.Create(ctx, exp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/experiments")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAPIEndExperiment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")

	exp, err := s.experimentRepo.GetByID(ctx, id)
	if err != nil {
		http.Error(w, "Experiment not found", http.StatusNotFound)
		return
	}
	if exp.EndedAt != nil {
		http.Error(w, "Experiment already ended", http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	exp.EndedAt = &now
	exp.IsActive = false

	if err := s.experimentRepo.Update(ctx, exp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/experiments")
	w.WriteHeader(http.StatusOK)
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
