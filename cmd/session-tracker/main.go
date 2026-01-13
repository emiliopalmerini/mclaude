package main

import (
	"encoding/json"
	"fmt"
	"os"

	"claude-watcher/internal/database"
	"claude-watcher/internal/tracker/adapters/logger"
	"claude-watcher/internal/tracker/adapters/prompter"
	"claude-watcher/internal/tracker/adapters/repository"
	"claude-watcher/internal/tracker/adapters/transcript"
	"claude-watcher/internal/tracker/domain"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(0)
}

func run() error {
	config, err := getConfig()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	log := logger.NewFileLogger(config.HomeDir)
	log.Debug("Session tracker started")

	// Read hook input from stdin
	var input domain.HookInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		log.Error(fmt.Sprintf("Failed to parse hook input: %v", err))
		return fmt.Errorf("parse hook input: %w", err)
	}
	log.Debug(fmt.Sprintf("Processing session %s", input.SessionID))

	// 1. Parse transcript (local file, fast)
	parser := transcript.NewParser()
	stats, err := parser.Parse(input.TranscriptPath)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to parse transcript: %v", err))
		return fmt.Errorf("parse transcript: %w", err)
	}
	log.Debug(fmt.Sprintf("Parsed stats: prompts=%d, responses=%d, tools=%d",
		stats.UserPrompts, stats.AssistantResponses, stats.ToolCalls))

	// 2. Collect quality feedback via TUI (no DB needed - this is the interactive part)
	bubbleTeaPrompter := prompter.NewBubbleTeaPrompter(log)
	qualityData, err := bubbleTeaPrompter.CollectQualityData(nil)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to collect quality data: %v", err))
		// Continue without quality data
	}

	// 3. Connect to database and save (after user interaction completes)
	db, err := database.NewTursoNoPing(config.DatabaseURL, config.AuthToken)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to connect to database: %v", err))
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	// Save session
	repo := repository.NewTursoRepository(db)
	instanceID := domain.GenerateInstanceID(config.Hostname, config.HomeDir)
	session := domain.NewSession(
		input.SessionID,
		instanceID,
		config.Hostname,
		input.ExitReason,
		input.PermissionMode,
		input.CWD,
		stats,
		qualityData,
	)

	if err := repo.Save(session); err != nil {
		log.Error(fmt.Sprintf("Failed to save session: %v", err))
		return fmt.Errorf("save session: %w", err)
	}

	if len(qualityData.Tags) > 0 {
		if err := repo.SaveTags(input.SessionID, qualityData.Tags); err != nil {
			log.Error(fmt.Sprintf("Failed to save tags: %v", err))
		}
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
		DatabaseURL: os.Getenv("CLAUDE_WATCHER_DATABASE_URL"),
		AuthToken:   os.Getenv("CLAUDE_WATCHER_AUTH_TOKEN"),
	}

	if cfg.DatabaseURL == "" {
		return cfg, fmt.Errorf("CLAUDE_WATCHER_DATABASE_URL not set")
	}

	if cfg.AuthToken == "" {
		return cfg, fmt.Errorf("CLAUDE_WATCHER_AUTH_TOKEN not set")
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
