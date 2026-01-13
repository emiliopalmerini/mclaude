package main

import (
	"claude-watcher/internal/database"
	"claude-watcher/internal/tracker/adapters/logger"
	"claude-watcher/internal/tracker/adapters/prompter"
	"claude-watcher/internal/tracker/adapters/repository"
	"claude-watcher/internal/tracker/adapters/transcript"
	"claude-watcher/internal/tracker/domain"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	// Exit cleanly on any error to not block session end
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(0)
	}
}

func run() error {
	// Get configuration from environment
	config, err := getConfig()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	// Initialize logger
	log := logger.NewFileLogger(config.HomeDir)
	log.Debug("Session tracker started")

	// Read hook input from stdin
	var input domain.HookInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		log.Error(fmt.Sprintf("Failed to parse hook input: %v", err))
		return fmt.Errorf("parse hook input: %w", err)
	}

	log.Debug(fmt.Sprintf("Processing session %s", input.SessionID))

	// Connect to database
	db, err := database.NewTurso(config.DatabaseURL, config.AuthToken)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to connect to database: %v", err))
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	// Initialize adapters
	parser := transcript.NewParser()
	repo := repository.NewTursoRepository(db)
	bubbleTeaPrompter := prompter.NewBubbleTeaPrompter(log)

	// Create service
	service := domain.NewService(parser, repo, bubbleTeaPrompter, log)

	// Track session
	instanceID := domain.GenerateInstanceID(config.Hostname, config.HomeDir)
	if err := service.TrackSession(input, instanceID, config.Hostname); err != nil {
		return err
	}

	log.Debug("Session tracking completed successfully")
	return nil
}

type config struct {
	DatabaseURL string
	AuthToken   string
	Hostname    string
	HomeDir     string
}

func getConfig() (config, error) {
	cfg := config{
		DatabaseURL: os.Getenv("TURSO_DATABASE_URL"),
		AuthToken:   os.Getenv("TURSO_AUTH_TOKEN"),
	}

	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("TURSO_DATABASE_URL not set")
	}

	if cfg.AuthToken == "" {
		return cfg, fmt.Errorf("TURSO_AUTH_TOKEN not set")
	}

	var err error
	cfg.Hostname, err = os.Hostname()
	if err != nil {
		return cfg, fmt.Errorf("get hostname: %w", err)
	}

	cfg.HomeDir, err = os.UserHomeDir()
	if err != nil {
		return cfg, fmt.Errorf("get home dir: %w", err)
	}

	return cfg, nil
}
