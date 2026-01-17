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
	SessionID       string
	MessageCountUser     int64
	MessageCountAssistant int64
	TurnCount       int64
	TokenInput      int64
	TokenOutput     int64
	TokenCacheRead  int64
	TokenCacheWrite int64
	CostEstimateUSD *float64
	ErrorCount      int64
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
