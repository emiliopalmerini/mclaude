package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	server := web.NewServer(
		app.DB.DB, servePort,
		app.ExperimentRepo, app.ExpVariableRepo, app.PricingRepo, app.SessionRepo, app.MetricsRepo, app.StatsRepo, app.ProjectRepo,
	)
	return server.Start(ctx)
}
