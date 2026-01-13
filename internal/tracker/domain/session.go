package domain

import "time"

// Session represents a Claude Code session with all its metadata and statistics
type Session struct {
	SessionID      string
	InstanceID     string
	Hostname       string
	Timestamp      time.Time
	ExitReason     string
	PermissionMode string
	WorkingDir     string
	Statistics     Statistics
	// Quality feedback (optional, collected at session end)
	Rating            *int     // 1-5 session satisfaction, nil if skipped
	PromptSpecificity *int     // 1-5 how detailed prompts were
	TaskCompletion    *int     // 1-5 how complete the work is
	CodeConfidence    *int     // 1-5 confidence in generated code
	Notes             string   // Free-form notes
	Tags              []string // Selected tag names (task_type only)
	// Limit tracking
	LimitMessage string // Captured if session ended due to hitting limit
}

// NewSession creates a new session with the given parameters
func NewSession(
	sessionID, instanceID, hostname, exitReason, permissionMode, workingDir string,
	stats Statistics,
	quality QualityData,
) Session {
	return Session{
		SessionID:         sessionID,
		InstanceID:        instanceID,
		Hostname:          hostname,
		Timestamp:         time.Now().UTC(),
		ExitReason:        exitReason,
		PermissionMode:    permissionMode,
		WorkingDir:        workingDir,
		Statistics:        stats,
		Rating:            quality.Rating,
		PromptSpecificity: quality.PromptSpecificity,
		TaskCompletion:    quality.TaskCompletion,
		CodeConfidence:    quality.CodeConfidence,
		Notes:             quality.Notes,
		Tags:              quality.Tags,
		LimitMessage:      stats.LimitMessage,
	}
}
