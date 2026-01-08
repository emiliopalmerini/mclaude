package sessions

import (
	"net/http"
	"strconv"

	"claude-watcher/internal/database/sqlc"
	"claude-watcher/internal/shared/middleware"
)

type Handler struct {
	queries *sqlc.Queries
}

func NewHandler(queries *sqlc.Queries) *Handler {
	return &Handler{queries: queries}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize := int64(20)
	offset := int64((page - 1)) * pageSize

	sessions, err := h.queries.ListSessions(ctx, sqlc.ListSessionsParams{
		Limit:  pageSize,
		Offset: offset,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, _ := h.queries.CountSessions(ctx)
	totalPages := int((count + pageSize - 1) / pageSize)

	data := SessionsData{
		Sessions:   sessions,
		Page:       page,
		TotalPages: totalPages,
	}

	if middleware.IsHTMX(r) {
		SessionsTable(data).Render(ctx, w)
		return
	}

	SessionsList(data).Render(ctx, w)
}
