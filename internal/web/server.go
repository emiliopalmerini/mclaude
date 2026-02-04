package web

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/adapters/prometheus"
	"github.com/emiliopalmerini/mclaude/internal/adapters/storage"
	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/ports"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	db                *sql.DB
	router            *http.ServeMux
	port              int
	transcriptStorage *storage.TranscriptStorage
	qualityRepo       *turso.SessionQualityRepository
	planConfigRepo    *turso.PlanConfigRepository
	promClient        ports.PrometheusClient
}

func NewServer(db *sql.DB, port int, ts *storage.TranscriptStorage) *Server {
	// Initialize Prometheus client (graceful degradation if not configured)
	var promClient ports.PrometheusClient
	promCfg := prometheus.LoadConfig()
	if client, err := prometheus.NewClient(promCfg); err == nil {
		promClient = client
	} else {
		promClient = prometheus.NewNoOpClient()
	}

	s := &Server{
		db:                db,
		router:            http.NewServeMux(),
		port:              port,
		transcriptStorage: ts,
		qualityRepo:       turso.NewSessionQualityRepository(db),
		planConfigRepo:    turso.NewPlanConfigRepository(db),
		promClient:        promClient,
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
		w.Write([]byte("ok"))
	})

	// Pages
	s.router.HandleFunc("GET /", s.handleDashboard)
	s.router.HandleFunc("GET /sessions", s.handleSessions)
	s.router.HandleFunc("GET /sessions/{id}", s.handleSessionDetail)
	s.router.HandleFunc("GET /sessions/{id}/review", s.handleSessionReview)
	s.router.HandleFunc("GET /experiments", s.handleExperiments)
	s.router.HandleFunc("GET /experiments/compare", s.handleExperimentCompare)
	s.router.HandleFunc("GET /experiments/{id}", s.handleExperimentDetail)
	s.router.HandleFunc("GET /settings", s.handleSettings)

	// API endpoints (for HTMX)
	s.router.HandleFunc("GET /api/stats", s.handleAPIStats)
	s.router.HandleFunc("GET /api/charts/tokens", s.handleAPIChartTokens)
	s.router.HandleFunc("GET /api/charts/cost", s.handleAPIChartCost)
	s.router.HandleFunc("POST /api/experiments", s.handleAPICreateExperiment)
	s.router.HandleFunc("POST /api/experiments/{id}/activate", s.handleAPIActivateExperiment)
	s.router.HandleFunc("POST /api/experiments/{id}/deactivate", s.handleAPIDeactivateExperiment)
	s.router.HandleFunc("DELETE /api/experiments/{id}", s.handleAPIDeleteExperiment)

	// Quality review endpoints
	s.router.HandleFunc("POST /api/sessions/{id}/quality", s.handleAPISaveQuality)

	// Real-time usage endpoint (for HTMX auto-refresh)
	s.router.HandleFunc("GET /api/realtime/usage", s.handleAPIRealtimeUsage)
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
