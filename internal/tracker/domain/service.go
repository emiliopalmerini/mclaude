package domain

import "fmt"

// HookInput represents input from Claude Code SessionEnd hook
type HookInput struct {
	SessionID      string
	TranscriptPath string
	ExitReason     string
	CWD            string
	PermissionMode string
}

// Service handles the business logic for tracking sessions
type Service struct {
	parser     TranscriptParser
	repository SessionRepository
	logger     Logger
}

// NewService creates a new tracker service
func NewService(
	parser TranscriptParser,
	repository SessionRepository,
	logger Logger,
) *Service {
	return &Service{
		parser:     parser,
		repository: repository,
		logger:     logger,
	}
}

// TrackSession processes a session end event and saves it
func (s *Service) TrackSession(input HookInput, instanceID, hostname string) error {
	s.logger.Debug(fmt.Sprintf("Processing session %s from %s", input.SessionID, input.TranscriptPath))

	// Parse transcript to extract statistics
	stats, err := s.parser.Parse(input.TranscriptPath)
	if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to parse transcript: %v", err))
		return fmt.Errorf("parse transcript: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("Parsed stats: prompts=%d, responses=%d, tools=%d, input_tokens=%d, output_tokens=%d",
		stats.UserPrompts, stats.AssistantResponses, stats.ToolCalls, stats.InputTokens, stats.OutputTokens))

	// Create session
	session := NewSession(
		input.SessionID,
		instanceID,
		hostname,
		input.ExitReason,
		input.PermissionMode,
		input.CWD,
		stats,
	)

	// Save session
	if err := s.repository.Save(session); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to save session: %v", err))
		return fmt.Errorf("save session: %w", err)
	}

	s.logger.Debug(fmt.Sprintf("Successfully saved session %s", input.SessionID))
	return nil
}
