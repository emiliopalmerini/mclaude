package domain

import "time"

type UsageMetric struct {
	ID         int64
	MetricName string
	Value      float64
	Attributes *string
	RecordedAt time.Time
	CreatedAt  time.Time
}

type UsageLimit struct {
	ID            string
	LimitValue    float64
	WarnThreshold float64
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type UsageSummary struct {
	TotalTokens float64
	TotalCost   float64
}

type PlanConfig struct {
	PlanType          string
	WindowHours       int
	LearnedTokenLimit *float64
	LearnedAt         *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Plan type constants
const (
	PlanPro    = "pro"
	PlanMax5x  = "max_5x"
	PlanMax20x = "max_20x"
	PlanCustom = "custom"
)

// Plan presets (messages per 5 hours - tokens not documented)
// These are rough estimates based on documented message limits
var PlanPresets = map[string]struct {
	Name            string
	MessagesPerWindow int
	TokenEstimate   float64 // Rough estimate: ~3K tokens per message avg
}{
	PlanPro:    {Name: "Pro", MessagesPerWindow: 45, TokenEstimate: 135000},
	PlanMax5x:  {Name: "Max 5x", MessagesPerWindow: 225, TokenEstimate: 675000},
	PlanMax20x: {Name: "Max 20x", MessagesPerWindow: 900, TokenEstimate: 2700000},
}

// Limit type constants (legacy - for manual limits)
const (
	LimitDailyTokens  = "daily_tokens"
	LimitWeeklyTokens = "weekly_tokens"
	LimitDailyCost    = "daily_cost"
	LimitWeeklyCost   = "weekly_cost"
)

// Metric name constants (from Claude Code OTEL)
const (
	MetricTokenUsage   = "claude_code.token.usage"
	MetricCostUsage    = "claude_code.cost.usage"
	MetricSessionCount = "claude_code.session.count"
)
