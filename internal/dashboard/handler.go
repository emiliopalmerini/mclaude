package dashboard

import (
	"net/http"

	"claude-watcher/internal/database/sqlc"
)

type Handler struct {
	queries *sqlc.Queries
}

func NewHandler(queries *sqlc.Queries) *Handler {
	return &Handler{queries: queries}
}

func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.queries.GetDashboardMetrics(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	today, _ := h.queries.GetTodayMetrics(ctx)
	week, _ := h.queries.GetWeekMetrics(ctx)

	data := DashboardData{
		Metrics: metrics,
		Today:   today,
		Week:    week,
	}

	Dashboard(data).Render(ctx, w)
}
