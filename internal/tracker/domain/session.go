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
	Rating *int     // 1-5 rating, nil if skipped
	Notes  string   // Free-form notes
	Tags   []string // Selected tag names
}

// NewSession creates a new session with the given parameters
func NewSession(
	sessionID, instanceID, hostname, exitReason, permissionMode, workingDir string,
	stats Statistics,
	quality QualityData,
) Session {
	return Session{
		SessionID:      sessionID,
		InstanceID:     instanceID,
		Hostname:       hostname,
		Timestamp:      time.Now().UTC(),
		ExitReason:     exitReason,
		PermissionMode: permissionMode,
		WorkingDir:     workingDir,
		Statistics:     stats,
		Rating:         quality.Rating,
		Notes:          quality.Notes,
		Tags:           quality.Tags,
	}
}
