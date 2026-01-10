package dashboard

import (
	"database/sql"
	"log"
	"net/http"

	apperrors "claude-watcher/internal/shared/errors"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.repo.GetDashboardMetrics(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	today, err := h.repo.GetTodayMetrics(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	week, err := h.repo.GetWeekMetrics(ctx)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	cacheMetrics, err := h.repo.GetCacheMetrics(ctx, "-7")
	if err != nil {
		log.Printf("error fetching cache metrics: %v", err)
	}

	topProject, err := h.repo.GetTopProject(ctx)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("error fetching top project: %v", err)
	}

	efficiencyMetrics, err := h.repo.GetEfficiencyMetrics(ctx, "-7")
	if err != nil {
		log.Printf("error fetching efficiency metrics: %v", err)
	}

	toolsBreakdown, err := h.repo.GetToolsBreakdownAll(ctx, "-7")
	if err != nil {
		log.Printf("error fetching tools breakdown: %v", err)
	}

	topTool := TopTool(toolsBreakdown)

	var cacheHitRate float64
	if totalTokens := toInt64(cacheMetrics.TotalTokens); totalTokens > 0 {
		cacheHitRate = float64(toInt64(cacheMetrics.CacheRead)) / float64(totalTokens) * 100
	}

	data := DashboardData{
		Metrics:           metrics,
		Today:             today,
		Week:              week,
		CacheMetrics:      cacheMetrics,
		TopProject:        topProject,
		EfficiencyMetrics: efficiencyMetrics,
		TopTool:           topTool,
		CacheHitRate:      cacheHitRate,
	}

	Dashboard(data).Render(ctx, w)
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case float64:
		return int64(val)
	default:
		return 0
	}
}
