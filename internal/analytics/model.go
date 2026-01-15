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
