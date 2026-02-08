package domain

type ToolEvent struct {
	ID           int64
	SessionID    string
	ToolName     string
	ToolUseID    string
	ToolInput    *string
	ToolResponse *string
	CapturedAt   string
}
