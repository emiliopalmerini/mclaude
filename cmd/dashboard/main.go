package main

import (
	"fmt"
	"os"

	"claude-watcher/internal/analytics"
	tursoRepo "claude-watcher/internal/analytics/outbound/turso"
	apptui "claude-watcher/internal/app/tui"
	"claude-watcher/internal/database"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	TursoDatabaseURL string `envconfig:"TURSO_DATABASE_URL_CLAUDE_WATCHER" required:"true"`
	TursoAuthToken   string `envconfig:"TURSO_AUTH_TOKEN_CLAUDE_WATCHER" required:"true"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Connect to database
	db, err := database.NewTurso(cfg.TursoDatabaseURL, cfg.TursoAuthToken)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	// Create analytics service
	repo := tursoRepo.NewRepository(db)
	logger := &consoleLogger{}
	service := analytics.NewService(repo, logger)

	// Create and run TUI
	app := apptui.NewApp(service)
	program := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		return fmt.Errorf("run dashboard: %w", err)
	}

	return nil
}

type consoleLogger struct{}

func (l *consoleLogger) Debug(msg string) {
	// Silent in production - could log to file if needed
}

func (l *consoleLogger) Error(msg string) {
	fmt.Fprintln(os.Stderr, msg)
}
