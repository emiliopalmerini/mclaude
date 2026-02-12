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
	// Filters
	FilterPeriod     string
	FilterExperiment string
	FilterProject    string
	Experiments      []FilterOption
	Projects         []FilterOption
}

// FilterOption for dropdown population.
type FilterOption struct {
	ID   string
	Name string
}

// SessionsPageData wraps session list with filter state.
type SessionsPageData struct {
	Sessions         []SessionSummary
	FilterExperiment string
	FilterProject    string
	FilterLimit      int
	Experiments      []FilterOption
	Projects         []FilterOption
	MaxTokens        int64
}

// SettingsPageData wraps pricing and plan config for the settings page.
type SettingsPageData struct {
	Pricing    []ModelPricing
	PlanConfig *PlanConfigView
}

// PlanConfigView for displaying plan config in settings.
type PlanConfigView struct {
	PlanType                string
	WindowHours             int
	LearnedTokenLimit       *float64
	WeeklyLearnedTokenLimit *float64
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

// RealtimeUsageStats contains real-time usage data for the JSON API.
type RealtimeUsageStats struct {
	Available       bool    `json:"available"`
	FiveHourTokens  float64 `json:"five_hour_tokens"`
	WeeklyTokens    float64 `json:"weekly_tokens"`
	FiveHourPercent float64 `json:"five_hour_percent"`
	WeeklyPercent   float64 `json:"weekly_percent"`
	FiveHourStatus  string  `json:"five_hour_status"`
	WeeklyStatus    string  `json:"weekly_status"`
	FiveHourLimit   float64 `json:"five_hour_limit"`
	WeeklyLimit     float64 `json:"weekly_limit"`
}

type ToolUsage struct {
	Name  string
	Count int64
}

type SessionSummary struct {
	ID             string
	ProjectID      string
	ProjectName    string
	ExperimentID   string
	ExperimentName string
	CreatedAt      string
	ExitReason     string
	Turns          int64
	Tokens         int64
	Cost           float64
	Model          string
	Duration       int64
	SubagentCount  int64
	// Quality (nil if not reviewed)
	IsReviewed    bool
	OverallRating int
	IsSuccess     *bool
}

type SubagentUsage struct {
	AgentType  string
	AgentKind  string // "task" or "skill"
	Count      int64
	Tokens     int64
	Cost       float64
	DurationMs int64
}

type ToolEventView struct {
	ToolName     string
	ToolUseID    string
	ToolInput    string
	ToolResponse string
	CapturedAt   string
}

type SessionDetail struct {
	ID                    string
	ProjectID             string
	ExperimentID          string
	ModelID               string
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
	Subagents             []SubagentUsage
	ToolEvents            []ToolEventView
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
	ModelID     string
	PlanType    string
	Notes       string
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
	ModelID     string
	PlanType    string
	Notes       string
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
	// Normalized behavior metrics
	TokensPerTurn    float64
	OutputRatio      float64
	CacheHitRate     float64
	ErrorRate        float64
	ToolCallsPerTurn float64
}

type ExperimentComparison struct {
	Experiments []ExperimentCompareItem
}

type ExperimentCompareItem struct {
	Name              string
	IsActive          bool
	ModelID           string
	PlanType          string
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
	// Normalized behavior metrics
	TokensPerTurn    float64
	OutputRatio      float64
	CacheHitRate     float64
	ErrorRate        float64
	ToolCallsPerTurn float64
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
