package web

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	db              *sql.DB
	router          *http.ServeMux
	port            int
	experimentRepo  ports.ExperimentRepository
	expVariableRepo ports.ExperimentVariableRepository
	pricingRepo     ports.PricingRepository
	sessionRepo     ports.SessionRepository
	metricsRepo     ports.SessionMetricsRepository
	statsRepo       ports.StatsRepository
	projectRepo     ports.ProjectRepository
}

func NewServer(
	db *sql.DB,
	port int,
	er ports.ExperimentRepository,
	evr ports.ExperimentVariableRepository,
	pr ports.PricingRepository,
	sr ports.SessionRepository,
	mr ports.SessionMetricsRepository,
	str ports.StatsRepository,
	projr ports.ProjectRepository,
) *Server {
	s := &Server{
		db:              db,
		router:          http.NewServeMux(),
		port:            port,
		experimentRepo:  er,
		expVariableRepo: evr,
		pricingRepo:     pr,
		sessionRepo:     sr,
		metricsRepo:     mr,
		statsRepo:       str,
		projectRepo:     projr,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(fmt.Sprintf("failed to create static filesystem: %v", err))
	}
	s.router.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Health check
	s.router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Pages
	s.router.HandleFunc("GET /", s.handleDashboard)
	s.router.HandleFunc("GET /sessions", s.handleSessions)
	s.router.HandleFunc("GET /sessions/{id}", s.handleSessionDetail)
	s.router.HandleFunc("GET /experiments", s.handleExperiments)
	s.router.HandleFunc("GET /experiments/compare", s.handleExperimentCompare)
	s.router.HandleFunc("GET /experiments/{id}", s.handleExperimentDetail)
	s.router.HandleFunc("GET /settings", s.handleSettings)

	// API endpoints (for HTMX)
	s.router.HandleFunc("GET /api/stats", s.handleAPIStats)
	s.router.HandleFunc("GET /api/charts/tokens", s.handleAPIChartTokens)
	s.router.HandleFunc("GET /api/charts/cost", s.handleAPIChartCost)
	s.router.HandleFunc("GET /api/charts/heatmap", s.handleAPIChartHeatmap)
	s.router.HandleFunc("POST /api/experiments", s.handleAPICreateExperiment)
	s.router.HandleFunc("POST /api/experiments/{id}/activate", s.handleAPIActivateExperiment)
	s.router.HandleFunc("POST /api/experiments/{id}/deactivate", s.handleAPIDeactivateExperiment)
	s.router.HandleFunc("DELETE /api/experiments/{id}", s.handleAPIDeleteExperiment)

	s.router.HandleFunc("POST /api/experiments/{id}/end", s.handleAPIEndExperiment)

	// Session management
	s.router.HandleFunc("DELETE /api/sessions/{id}", s.handleAPIDeleteSession)
	s.router.HandleFunc("POST /api/sessions/cleanup", s.handleAPICleanupSessions)

	// Pricing management
	s.router.HandleFunc("POST /api/pricing", s.handleAPICreatePricing)
	s.router.HandleFunc("POST /api/pricing/{id}/default", s.handleAPISetDefaultPricing)
	s.router.HandleFunc("DELETE /api/pricing/{id}", s.handleAPIDeletePricing)

	// Export
	s.router.HandleFunc("GET /api/export/sessions", s.handleAPIExportSessions)

}

func (s *Server) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Printf("Starting server at http://localhost:%d\n", s.port)

	// Handle graceful shutdown
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("Server shutdown error: %v\n", err)
		}
	}()

	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil // Graceful shutdown
	}
	return err
}
