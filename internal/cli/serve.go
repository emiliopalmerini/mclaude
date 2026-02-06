package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/adapters/prometheus"
	"github.com/emiliopalmerini/mclaude/internal/adapters/storage"
	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/ports"
	"github.com/emiliopalmerini/mclaude/internal/web"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the web dashboard",
	Long: `Start the local web dashboard server.

Examples:
  mclaude serve              # Start on default port 8080
  mclaude serve --port 3000  # Start on port 3000`,
	RunE: runServe,
}

var servePort int

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 8080, "Port to listen on")
}

func runServe(cmd *cobra.Command, args []string) error {
	db, err := turso.NewDB()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	transcriptStorage, err := storage.NewTranscriptStorage()
	if err != nil {
		return fmt.Errorf("failed to initialize transcript storage: %w", err)
	}

	qualityRepo := turso.NewSessionQualityRepository(db.DB)
	planConfigRepo := turso.NewPlanConfigRepository(db.DB)

	var promClient ports.PrometheusClient
	promCfg := prometheus.LoadConfig()
	if client, err := prometheus.NewClient(promCfg); err == nil {
		promClient = client
	} else {
		promClient = prometheus.NewNoOpClient()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	server := web.NewServer(db.DB, servePort, transcriptStorage, qualityRepo, planConfigRepo, promClient)
	return server.Start(ctx)
}
