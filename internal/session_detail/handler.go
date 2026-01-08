package session_detail

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"

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
	sessionID := chi.URLParam(r, "sessionID")

	session, err := h.queries.GetSessionByID(ctx, sessionID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	SessionDetail(session).Render(ctx, w)
}
