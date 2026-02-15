package web

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/util"
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
	_ = json.NewEncoder(w).Encode(stats)
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

	stats, err := queries.GetDailyStats(ctx, sqlc.GetDailyStatsParams{
		CreatedAt: startDate,
		Limit:     30,
	})
	if err != nil {
		slog.Error("api: chart tokens", "error", err)
		http.Error(w, "failed to fetch token stats", http.StatusInternalServerError)
		return
	}

	labels := make([]string, len(stats))
	tokens := make([]int64, len(stats))
	sessions := make([]int64, len(stats))

	for i, stat := range stats {
		labels[i] = formatChartDate(stat.Date)
		tokens[i] = util.ToInt64(stat.TotalTokens)
		sessions[i] = stat.SessionCount
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
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

	stats, err := queries.GetDailyStats(ctx, sqlc.GetDailyStatsParams{
		CreatedAt: startDate,
		Limit:     30,
	})
	if err != nil {
		slog.Error("api: chart cost", "error", err)
		http.Error(w, "failed to fetch cost stats", http.StatusInternalServerError)
		return
	}

	labels := make([]string, len(stats))
	costs := make([]float64, len(stats))

	for i, stat := range stats {
		labels[i] = formatChartDate(stat.Date)
		costs[i] = util.ToFloat64(stat.TotalCost)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"labels": labels,
		"costs":  costs,
	})
}

func (s *Server) handleAPIChartHeatmap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queries := sqlc.New(s.db)

	// Get daily stats for the current year
	startDate := time.Now().AddDate(0, 0, -365).Format(time.RFC3339)
	stats, err := queries.GetDailyStats(ctx, sqlc.GetDailyStatsParams{
		CreatedAt: startDate,
		Limit:     366,
	})
	if err != nil {
		slog.Error("api: chart heatmap", "error", err)
		http.Error(w, "failed to fetch heatmap stats", http.StatusInternalServerError)
		return
	}

	data := make([][2]any, len(stats))
	for i, stat := range stats {
		data[i] = [2]any{formatChartDate(stat.Date), stat.SessionCount}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": data,
	})
}
