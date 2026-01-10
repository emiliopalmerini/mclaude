package productivity

import (
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

	rangeParam := r.URL.Query().Get("range")
	if rangeParam == "" {
		rangeParam = "30d"
	}

	days := rangeToDays(rangeParam)
	hours := rangeToHours(rangeParam)

	efficiency, err := h.repo.GetEfficiencyMetrics(ctx, days)
	if err != nil {
		apperrors.HandleError(w, err)
		return
	}

	dailyTrends, err := h.repo.GetEfficiencyMetricsDaily(ctx, days)
	if err != nil {
		log.Printf("error fetching daily trends: %v", err)
	}

	dayOfWeek, err := h.repo.GetDayOfWeekDistribution(ctx, days)
	if err != nil {
		log.Printf("error fetching day of week distribution: %v", err)
	}

	hourOfDay, err := h.repo.GetHourOfDayDistribution(ctx, hours)
	if err != nil {
		log.Printf("error fetching hour of day distribution: %v", err)
	}

	toolsBreakdown, err := h.repo.GetToolsBreakdownAll(ctx, days)
	if err != nil {
		log.Printf("error fetching tools breakdown: %v", err)
	}

	topTools := AggregateTools(toolsBreakdown, 10)

	data := ProductivityData{
		Efficiency:  efficiency,
		DailyTrends: dailyTrends,
		DayOfWeek:   dayOfWeek,
		HourOfDay:   hourOfDay,
		TopTools:    topTools,
		Range:       rangeParam,
	}

	Productivity(data).Render(ctx, w)
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

func rangeToHours(r string) string {
	switch r {
	case "7d":
		return "-168"
	case "30d":
		return "-720"
	case "90d":
		return "-2160"
	default:
		return "-720"
	}
}
