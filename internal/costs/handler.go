package costs

import (
	"log"
	"net/http"

	apperrors "claude-watcher/internal/shared/errors"
)

const tokenCostPerMillion = 3.0 // Approximate cost per million input tokens

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "30d"
	}

	days := rangeToDays(rangeParam)

	projects, err := h.repo.GetProjectMetrics(ctx, days)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	models, err := h.repo.GetModelEfficiency(ctx, days)
	if err != nil {
		log.Printf("error fetching model efficiency: %v", err)
	}

	cacheMetrics, err := h.repo.GetCacheMetrics(ctx, days)
	if err != nil {
		log.Printf("error fetching cache metrics: %v", err)
	}

	cacheDaily, err := h.repo.GetCacheMetricsDaily(ctx, days)
	if err != nil {
		log.Printf("error fetching cache daily metrics: %v", err)
	}

	var cacheHitRate float64
	var totalSavings float64
	cacheRead := asInt64(cacheMetrics.CacheRead)
	totalTokens := asInt64(cacheMetrics.TotalTokens)
	if cacheRead+totalTokens > 0 {
		cacheHitRate = float64(cacheRead) / float64(cacheRead+totalTokens) * 100
		totalSavings = float64(cacheRead) / 1_000_000 * tokenCostPerMillion
	}

	data := CostsData{
		Projects:     projects,
		Models:       models,
		CacheMetrics: cacheMetrics,
		CacheDaily:   cacheDaily,
		TotalSavings: totalSavings,
		CacheHitRate: cacheHitRate,
		Range:        rangeParam,
	}

	Costs(data).Render(ctx, w)
}

func rangeToDays(r string) string {
	switch r {
	case "7d":
		return "-7"
	case "30d":
		return "-30"
	case "90d":
		return "-90"
	default:
		return "-30"
	}
}

func asInt64(v interface{}) int64 {
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
