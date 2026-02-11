package domain

import "time"

type Session struct {
	ID                   string
	ProjectID            string
	ExperimentID         *string
	TranscriptPath       string
	TranscriptStoredPath *string
	Cwd                  string
	PermissionMode       string
	ExitReason           string
	StartedAt            *time.Time
	EndedAt              *time.Time
	DurationSeconds      *int64
	CreatedAt            time.Time
}

type SessionMetrics struct {
	SessionID             string
	ModelID               *string // e.g., "claude-opus-4-5-20251101"
	MessageCountUser      int64
	MessageCountAssistant int64
	TurnCount             int64
	TokenInput            int64
	TokenOutput           int64
	TokenCacheRead        int64
	TokenCacheWrite       int64
	CostEstimateUSD       *float64
	ErrorCount            int64
	InputRate             *float64
	OutputRate            *float64
	CacheReadRate         *float64
	CacheWriteRate        *float64
}

type SessionTool struct {
	ID              int64
	SessionID       string
	ToolName        string
	InvocationCount int64
	TotalDurationMs *int64
	ErrorCount      int64
}

type SessionFile struct {
	ID             int64
	SessionID      string
	FilePath       string
	Operation      string // "read", "write", "edit"
	OperationCount int64
}

type SessionCommand struct {
	ID         int64
	SessionID  string
	Command    string
	ExitCode   *int
	ExecutedAt *time.Time
}

// SessionListItem is a pre-joined summary for listing sessions with metrics.
type SessionListItem struct {
	ID            string
	ProjectID     string
	ExperimentID  *string
	ExitReason    string
	CreatedAt     string
	Duration      *int64
	TurnCount     int64
	TotalTokens   int64
	Cost          *float64
	ModelID       *string
	SubagentCount int64
}

type SessionSubagent struct {
	ID              int64
	SessionID       string
	AgentType       string   // subagent_type for Task (e.g. "Explore", "Bash"), skill name for Skill (e.g. "commit")
	AgentKind       string   // "task" or "skill"
	Description     *string  // short description from Task input
	Model           *string  // model alias (e.g. "haiku", "sonnet") or nil
	TotalTokens     int64
	TokenInput      int64
	TokenOutput     int64
	TokenCacheRead  int64
	TokenCacheWrite int64
	TotalDurationMs *int64
	ToolUseCount    int64
	CostEstimateUSD *float64
}
