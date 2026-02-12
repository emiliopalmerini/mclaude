package domain

// AggregateStats holds summary statistics across sessions.
type AggregateStats struct {
	SessionCount           int64
	TotalUserMessages      int64
	TotalAssistantMessages int64
	TotalTurns             int64
	TotalTokenInput        int64
	TotalTokenOutput       int64
	TotalTokenCacheRead    int64
	TotalTokenCacheWrite   int64
	TotalCostUsd           float64
	TotalErrors            int64
}

// ToolUsageStats holds usage data for a single tool.
type ToolUsageStats struct {
	ToolName         string
	TotalInvocations int64
	TotalErrors      int64
}

// ExperimentStats holds aggregate stats for a specific experiment.
type ExperimentStats struct {
	ExperimentID   string
	ExperimentName string
	AggregateStats
}

// NormalizedMetrics holds pricing-independent behavioral metrics for experiment comparison.
type NormalizedMetrics struct {
	TokensPerTurn    float64
	OutputRatio      float64
	CacheHitRate     float64
	ErrorRate        float64
	ToolCallsPerTurn float64
}

// ComputeNormalized derives behavioral metrics from aggregate stats.
// All divisions are zero-safe: returns 0 when the divisor is zero.
func (a *AggregateStats) ComputeNormalized(totalToolCalls int64) NormalizedMetrics {
	var m NormalizedMetrics

	if a.TotalTurns > 0 {
		m.TokensPerTurn = float64(a.TotalTokenInput+a.TotalTokenOutput) / float64(a.TotalTurns)
		m.ErrorRate = float64(a.TotalErrors) / float64(a.TotalTurns)
		m.ToolCallsPerTurn = float64(totalToolCalls) / float64(a.TotalTurns)
	}

	if a.TotalTokenInput > 0 {
		m.OutputRatio = float64(a.TotalTokenOutput) / float64(a.TotalTokenInput)
	}

	contextTotal := a.TotalTokenInput + a.TotalTokenCacheRead + a.TotalTokenCacheWrite
	if contextTotal > 0 {
		m.CacheHitRate = float64(a.TotalTokenCacheRead) / float64(contextTotal)
	}

	return m
}

// TranscriptPathInfo holds session ID and transcript path for cleanup operations.
type TranscriptPathInfo struct {
	ID             string
	TranscriptPath string
}
