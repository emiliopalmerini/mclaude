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

// SettingsPageData wraps pricing for the settings page.
type SettingsPageData struct {
	Pricing []ModelPricing
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

type ExperimentVariable struct {
	Key   string
	Value string
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
	Variables   []ExperimentVariable
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
	Variables         []ExperimentVariable
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
