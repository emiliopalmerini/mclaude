package templates

type DashboardStats struct {
	SessionCount     int64
	TotalTokens      int64
	TotalCost        float64
	TotalTurns       int64
	TokenInput       int64
	TokenOutput      int64
	CacheRead        int64
	CacheWrite       int64
	TotalErrors      int64
	ActiveExperiment string
	DefaultModel     string // Display name of the default model for cost calculations
	TopTools         []ToolUsage
	RecentSessions   []SessionSummary
	// Quality stats
	ReviewedCount int64
	SuccessRate   *float64
	AvgOverall    *float64
	// Usage limits
	UsageStats *UsageLimitStats
}

type UsageLimitStats struct {
	PlanType     string  // pro, max_5x, max_20x
	WindowHours  int     // Rolling window size (typically 5)
	TokensUsed   float64 // Current tokens in window
	TokenLimit   float64 // Limit (learned or estimated)
	UsagePercent float64 // 0-100+
	Status       string  // OK, WARNING, EXCEEDED
	IsLearned    bool    // true if limit is learned, false if estimated
	MinutesLeft  int     // Minutes until window resets (approx)
	// Weekly fields
	WeeklyTokensUsed   float64 // Current tokens in weekly window
	WeeklyTokenLimit   float64 // Weekly limit (learned or estimated)
	WeeklyUsagePercent float64 // 0-100+
	WeeklyStatus       string  // OK, WARNING, EXCEEDED
	WeeklyIsLearned    bool    // true if weekly limit is learned
}

// RealtimeUsageStats contains real-time usage from Prometheus.
type RealtimeUsageStats struct {
	Available        bool    // true if Prometheus data is available
	Source           string  // "prometheus" or "local"
	FiveHourTokens   float64 // Tokens in 5-hour window
	FiveHourCost     float64 // Cost in 5-hour window
	WeeklyTokens     float64 // Tokens in 7-day window
	WeeklyCost       float64 // Cost in 7-day window
	FiveHourPercent  float64 // 0-100+ based on limit
	WeeklyPercent    float64 // 0-100+ based on limit
	FiveHourStatus   string  // OK, WARNING, EXCEEDED
	WeeklyStatus     string  // OK, WARNING, EXCEEDED
	FiveHourLimit    float64 // Token limit for 5-hour window
	WeeklyLimit      float64 // Token limit for weekly window
}

type ToolUsage struct {
	Name  string
	Count int64
}

type SessionSummary struct {
	ID           string
	ProjectID    string
	ExperimentID string
	CreatedAt    string
	ExitReason   string
	Turns        int64
	Tokens       int64
	Cost         float64
	// Quality (nil if not reviewed)
	IsReviewed    bool
	OverallRating int
	IsSuccess     *bool
}

type SessionDetail struct {
	ID                    string
	ProjectID             string
	ExperimentID          string
	Cwd                   string
	PermissionMode        string
	ExitReason            string
	StartedAt             string
	EndedAt               string
	DurationSeconds       int64
	CreatedAt             string
	MessageCountUser      int64
	MessageCountAssistant int64
	TurnCount             int64
	TokenInput            int64
	TokenOutput           int64
	TokenCacheRead        int64
	TokenCacheWrite       int64
	CostEstimateUsd       float64
	ErrorCount            int64
	Tools                 []ToolUsage
	Files                 []FileOperation
	// Quality
	Quality *SessionQuality
}

type FileOperation struct {
	Path      string
	Operation string
	Count     int64
}

type Experiment struct {
	ID          string
	Name        string
	Description string
	Hypothesis  string
	StartedAt   string
	EndedAt     string
	IsActive    bool
	CreatedAt   string
	// Stats
	SessionCount   int64
	TotalTokens    int64
	TotalCost      float64
	TokensPerSess  int64
	CostPerSession float64
}

type ExperimentDetail struct {
	ID          string
	Name        string
	Description string
	Hypothesis  string
	StartedAt   string
	EndedAt     string
	IsActive    bool
	CreatedAt   string
	// Stats
	SessionCount      int64
	TotalTurns        int64
	UserMessages      int64
	AssistantMessages int64
	TotalErrors       int64
	TokenInput        int64
	TokenOutput       int64
	CacheRead         int64
	CacheWrite        int64
	TotalTokens       int64
	TotalCost         float64
	TokensPerSession  int64
	CostPerSession    float64
	// Top tools
	TopTools []ToolUsage
	// Recent sessions
	RecentSessions []SessionSummary
	// Quality stats
	ReviewedCount  int64
	AvgOverall     *float64
	SuccessRate    *float64
	AvgAccuracy    *float64
	AvgHelpfulness *float64
	AvgEfficiency  *float64
}

type ExperimentComparison struct {
	Experiments []ExperimentCompareItem
}

type ExperimentCompareItem struct {
	Name              string
	IsActive          bool
	SessionCount      int64
	TotalTurns        int64
	UserMessages      int64
	AssistantMessages int64
	TotalErrors       int64
	TokenInput        int64
	TokenOutput       int64
	CacheRead         int64
	CacheWrite        int64
	TotalTokens       int64
	TotalCost         float64
	TokensPerSession  int64
	CostPerSession    float64
	// Quality metrics
	ReviewedCount  int64
	AvgOverall     *float64
	SuccessRate    *float64
	AvgAccuracy    *float64
	AvgHelpfulness *float64
	AvgEfficiency  *float64
}

type ModelPricing struct {
	ID                   string
	DisplayName          string
	InputPerMillion      float64
	OutputPerMillion     float64
	CacheReadPerMillion  float64
	CacheWritePerMillion float64
	IsDefault            bool
}

// SessionQuality for review form and display
type SessionQuality struct {
	SessionID         string
	OverallRating     int // 0 means unset
	IsSuccess         *bool
	AccuracyRating    int
	HelpfulnessRating int
	EfficiencyRating  int
	Notes             string
	ReviewedAt        string
}

// TranscriptMessage for transcript viewer
type TranscriptMessage struct {
	Role      string
	Content   string
	Timestamp string
	Tools     []TranscriptToolUse
}

// TranscriptToolUse for displaying tool invocations
type TranscriptToolUse struct {
	Name  string
	Input string
}

// SessionReviewData combines session detail with quality and transcript
type SessionReviewData struct {
	SessionDetail
	Quality    SessionQuality
	Transcript []TranscriptMessage
}

// QualityStats for experiment comparison
type QualityStats struct {
	ReviewedCount  int64
	AvgOverall     *float64
	SuccessCount   int64
	FailureCount   int64
	SuccessRate    *float64
	AvgAccuracy    *float64
	AvgHelpfulness *float64
	AvgEfficiency  *float64
}
