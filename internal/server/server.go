package server

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"claude-watcher/internal/dashboard"
	"claude-watcher/internal/database/sqlc"
	"claude-watcher/internal/session_detail"
	"claude-watcher/internal/sessions"
	sharedmw "claude-watcher/internal/shared/middleware"
)

func NewHTTPServer(addr string, db *sql.DB) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(sharedmw.HTMX)

	r.Handle("/static/*", http.StripPrefix("/static/",
		http.FileServer(http.Dir("static"))))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})

	queries := sqlc.New(db)

	dashboard.RegisterRoutes(r, dashboard.NewHandler(queries))
	sessions.RegisterRoutes(r, sessions.NewHandler(queries))
	session_detail.RegisterRoutes(r, session_detail.NewHandler(queries))

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}
