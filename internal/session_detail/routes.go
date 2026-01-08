package session_detail

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/sessions/{sessionID}", h.Show)
}
