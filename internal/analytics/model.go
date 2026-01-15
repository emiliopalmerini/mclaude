package analytics

import "time"

// OverviewMetrics contains aggregate metrics for the dashboard overview
type OverviewMetrics struct {
	TotalSessions int
	TotalCost     float64
	Tokens        TokenSummary
	LimitHits     int
	LastLimitHit  *time.Time
}

// TokenSummary aggregates token usage
type TokenSummary struct {
	Input      int64
	Output     int64
	Thinking   int64
	CacheRead  int64
	CacheWrite int64
}

// Total returns the sum of input and output tokens
func (t TokenSummary) Total() int64 {
	return t.Input + t.Output
}

// SessionSummary is a lightweight session for list views
type SessionSummary struct {
	SessionID     string
	Timestamp     time.Time
	WorkingDir    string
	Model         string
	EstimatedCost float64
	TotalTokens   int64
	ToolCalls     int
	Rating        *int
}

// SessionFilter defines criteria for filtering sessions
type SessionFilter struct {
	Limit  int
	Offset int
}

// SessionDetail contains full session data for detail view
type SessionDetail struct {
	SessionID          string
	Hostname           string
	Timestamp          time.Time
	ExitReason         string
	WorkingDirectory   string
	GitBranch          string
	Model              string
	ClaudeVersion      string
	DurationSeconds    int
	Tokens             TokenSummary
	EstimatedCost      float64
	UserPrompts        int
	AssistantResponses int
	ToolCalls          int
	ErrorsCount        int
	Summary            string
	Rating             *int
	PromptSpecificity  *int
	TaskCompletion     *int
	CodeConfidence     *int
	Notes              string
}

// CostsBreakdown contains cost analysis data
type CostsBreakdown struct {
	TotalCost  float64
	TodayCost  float64
	WeekCost   float64
	ByModel    []ModelCostRow
	DailyTrend []DailyCost
	ByProject  []ProjectCostRow
}

// ModelCostRow represents cost data for a single model
type ModelCostRow struct {
	Model              string
	Sessions           int
	Cost               float64
	CostPerMillionToks float64
}

// DailyCost represents cost for a single day
type DailyCost struct {
	Date string
	Cost float64
}

// ProjectCostRow represents cost data for a single project
type ProjectCostRow struct {
	Project  string
	Sessions int
	Cost     float64
}
