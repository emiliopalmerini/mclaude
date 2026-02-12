package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	_ "github.com/tursodatabase/go-libsql"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/migrate"
	"github.com/emiliopalmerini/mclaude/internal/web"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	url := os.Getenv("MCLAUDE_DATABASE_URL")
	if url == "" {
		return fmt.Errorf("MCLAUDE_DATABASE_URL is required")
	}
	token := os.Getenv("MCLAUDE_AUTH_TOKEN")
	if token == "" {
		return fmt.Errorf("MCLAUDE_AUTH_TOKEN is required")
	}

	port := 8080
	if p := os.Getenv("PORT"); p != "" {
		var err error
		port, err = strconv.Atoi(p)
		if err != nil {
			return fmt.Errorf("invalid PORT: %s", p)
		}
	}

	db, err := turso.NewRemoteDB(url, token)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	if err := migrate.RunAll(ctx, db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	repos := turso.NewRepositories(db)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	server := web.NewServer(
		db, port, repos.Transcripts,
		repos.Quality, repos.Experiments,
		repos.ExperimentVariables, repos.Pricing, repos.Sessions, repos.Metrics,
		repos.Stats, repos.Projects,
	)
	return server.Start(ctx)
}
