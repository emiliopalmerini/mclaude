package domain

import "fmt"

// HookInput represents input from Claude Code SessionEnd hook
type HookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	ExitReason     string `json:"reason"`
	CWD            string `json:"cwd"`
	PermissionMode string `json:"permission_mode"`
	IsInteractive  bool   `json:"is_interactive"`
}

// Service handles the business logic for tracking sessions
type Service struct {
	parser     TranscriptParser
	repository SessionRepository
	prompter   Prompter
	logger     Logger
}

// NewService creates a new tracker service
func NewService(
	parser TranscriptParser,
	repository SessionRepository,
	prompter Prompter,
	logger Logger,
) *Service {
	return &Service{
		parser:     parser,
		repository: repository,
		prompter:   prompter,
		logger:     logger,
	}
}

// TrackSession processes a session end event and saves it
func (s *Service) TrackSession(input HookInput, instanceID, hostname string) error {
	s.logger.Debug(fmt.Sprintf("Processing session %s from %s", input.SessionID, input.TranscriptPath))

	// Parse transcript to extract statistics and limit events
	parsed, err := s.parser.Parse(input.TranscriptPath)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to parse transcript: %v", err))
		return fmt.Errorf("parse transcript: %w", err)
	}

	stats := parsed.Statistics
	s.logger.Debug(fmt.Sprintf("Parsed stats: prompts=%d, responses=%d, tools=%d, input_tokens=%d, output_tokens=%d, limit_events=%d",
		stats.UserPrompts, stats.AssistantResponses, stats.ToolCalls, stats.InputTokens, stats.OutputTokens, len(parsed.LimitEvents)))

	// Collect quality feedback from user (optional)
	var qualityData QualityData
	if s.prompter != nil {
		var tags []Tag
		var err error
		tags, err = s.repository.GetAllTags()
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to get tags: %v", err))
			// Continue with empty tags - non-fatal error
			tags = []Tag{}
		}
		qualityData, err = s.prompter.CollectQualityData(tags)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Failed to collect quality data: %v", err))
			// Continue without quality data - non-fatal error
		}
	}

	// Create session
	session := NewSession(
		input.SessionID,
		instanceID,
		hostname,
		input.ExitReason,
		input.PermissionMode,
		input.CWD,
		stats,
		qualityData,
	)

	// Save session
	if err := s.repository.Save(session); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to save session: %v", err))
		return fmt.Errorf("save session: %w", err)
	}

	// Save tags (if any)
	if len(qualityData.Tags) > 0 {
		if err := s.repository.SaveTags(input.SessionID, qualityData.Tags); err != nil {
			s.logger.Error(fmt.Sprintf("Failed to save tags: %v", err))
			// Continue - session was saved successfully
		}
	}

	s.logger.Debug(fmt.Sprintf("Successfully saved session %s", input.SessionID))
	return nil
}
