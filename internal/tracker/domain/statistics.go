package domain

import "time"

// Statistics holds all metrics for a Claude Code session
type Statistics struct {
	// Interaction metrics
	UserPrompts        int
	AssistantResponses int
	ToolCalls          int
	ToolsBreakdown     map[string]int
	ErrorsCount        int

	// File tracking
	FilesAccessed []string
	FilesModified []string

	// Token usage
	InputTokens      int
	OutputTokens     int
	ThinkingTokens   int
	CacheReadTokens  int
	CacheWriteTokens int

	// Session metadata
	Model         string
	GitBranch     string
	ClaudeVersion string
	Summary       string
	StartTime     *time.Time
	EndTime       *time.Time

	// Limit tracking
	LimitMessage string
}

// Duration calculates session duration in seconds
func (s Statistics) Duration() int {
	if s.StartTime == nil || s.EndTime == nil {
		return 0
	}
	return int(s.EndTime.Sub(*s.StartTime).Seconds())
}

// NewStatistics creates an empty Statistics instance
func NewStatistics() Statistics {
	return Statistics{
		ToolsBreakdown: make(map[string]int),
		FilesAccessed:  []string{},
		FilesModified:  []string{},
	}
}
