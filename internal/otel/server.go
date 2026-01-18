package otel

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
)

type Server struct {
	db          *sql.DB
	metricsRepo *turso.UsageMetricsRepository
	port        int
	healthPort  int
}

func NewServer(db *sql.DB, port, healthPort int) *Server {
	return &Server{
		db:          db,
		metricsRepo: turso.NewUsageMetricsRepository(db),
		port:        port,
		healthPort:  healthPort,
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/metrics", s.handleMetrics)
	mux.HandleFunc("GET /health", s.handleHealth)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	log.Printf("Starting OTLP receiver on port %d", s.port)
	return server.ListenAndServe()
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")

	var receiver *Receiver
	switch contentType {
	case "application/x-protobuf":
		receiver = NewReceiver(s.metricsRepo)
	case "application/json":
		receiver = NewReceiver(s.metricsRepo)
	default:
		// Default to protobuf
		receiver = NewReceiver(s.metricsRepo)
	}

	if err := receiver.HandleRequest(r.Context(), r.Body, contentType); err != nil {
		log.Printf("Error processing metrics: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
