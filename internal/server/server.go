package server

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"claude-watcher/internal/api"
	"claude-watcher/internal/costs"
	"claude-watcher/internal/dashboard"
	"claude-watcher/internal/database/sqlc"
	"claude-watcher/internal/productivity"
	"claude-watcher/internal/session_detail"
	"claude-watcher/internal/sessions"
	sharedmw "claude-watcher/internal/shared/middleware"
)

// Config holds server-specific configuration.
type Config struct {
	Addr              string
	DefaultPageSize   int64
	DefaultRangeHours int
}

func NewHTTPServer(cfg Config, db *sql.DB) *http.Server {
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

	// Create repositories
	dashboardRepo := dashboard.NewSQLCRepository(queries)
	sessionsRepo := sessions.NewSQLCRepository(queries)
	sessionDetailRepo := session_detail.NewSQLCRepository(queries)
	apiRepo := api.NewSQLCRepository(queries)
	productivityRepo := productivity.NewSQLCRepository(queries)
	costsRepo := costs.NewSQLCRepository(queries)

	// Register routes with handlers
	dashboard.RegisterRoutes(r, dashboard.NewHandler(dashboardRepo))
	sessions.RegisterRoutes(r, sessions.NewHandler(sessionsRepo, cfg.DefaultPageSize))
	session_detail.RegisterRoutes(r, session_detail.NewHandler(sessionDetailRepo))
	api.RegisterRoutes(r, api.NewHandler(apiRepo, cfg.DefaultRangeHours))
	productivity.RegisterRoutes(r, productivity.NewHandler(productivityRepo))
	costs.RegisterRoutes(r, costs.NewHandler(costsRepo))

	return &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}
}
