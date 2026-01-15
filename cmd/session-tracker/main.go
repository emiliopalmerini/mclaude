package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"claude-watcher/internal/database"
	"claude-watcher/internal/limits"
	limitsOutbound "claude-watcher/internal/limits/outbound"
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

	// 1. Parse transcript (local file, fast) - uses Phase 2 parser with limit event extraction
	parser := transcript.NewParser()
	parsed, err := parser.Parse(input.TranscriptPath)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to parse transcript: %v", err))
		return fmt.Errorf("parse transcript: %w", err)
	}
	stats := parsed.Statistics
	log.Debug(fmt.Sprintf("Parsed stats: prompts=%d, responses=%d, tools=%d, limit_events=%d",
		stats.UserPrompts, stats.AssistantResponses, stats.ToolCalls, len(parsed.LimitEvents)))

	// 2. Collect quality feedback via TUI (only if session was interactive)
	var qualityData domain.QualityData
	if input.IsInteractive {
		bubbleTeaPrompter := prompter.NewBubbleTeaPrompter(log)
		qualityData, err = bubbleTeaPrompter.CollectQualityData(domain.DefaultTaskTypeTags())
		if err != nil {
			log.Error(fmt.Sprintf("Failed to collect quality data: %v", err))
			// Continue without quality data
		}
	} else {
		log.Debug("Skipping quality prompts - session was not interactive")
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

	// Record limit events (if any were extracted from transcript)
	if len(parsed.LimitEvents) > 0 {
		limitsRepo := limitsOutbound.NewTursoRepository(db)
		limitsSvc := limits.NewService(limitsRepo, log)

		for _, event := range parsed.LimitEvents {
			// Only record "hit" events (not resets)
			if event.EventType != domain.LimitEventHit {
				continue
			}

			info := limits.ParsedLimitInfo{
				LimitType: convertLimitType(event.LimitType),
				Timestamp: parseTimestamp(event.Timestamp),
				Message:   event.Message,
			}

			if err := limitsSvc.RecordLimitHit(info); err != nil {
				log.Error(fmt.Sprintf("Failed to record limit event: %v", err))
			}
		}
		log.Debug(fmt.Sprintf("Processed %d limit events", len(parsed.LimitEvents)))
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

func convertLimitType(lt domain.LimitType) limits.LimitType {
	switch lt {
	case domain.LimitTypeDaily:
		return limits.LimitTypeDaily
	case domain.LimitTypeWeekly:
		return limits.LimitTypeWeekly
	case domain.LimitTypeMonthly:
		return limits.LimitTypeMonthly
	default:
		return limits.LimitTypeDaily
	}
}

func parseTimestamp(ts string) time.Time {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Now().UTC()
	}
	return t
}
