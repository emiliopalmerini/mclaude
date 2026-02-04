package ports

import (
	"context"
	"time"
)

// MetricsExporter exports session metrics to an external observability system.
type MetricsExporter interface {
	// ExportSessionMetrics exports enriched metrics for a completed session.
	ExportSessionMetrics(ctx context.Context, m *EnrichedMetrics) error
	// Close shuts down the exporter and flushes any pending metrics.
	Close(ctx context.Context) error
}

// EnrichedMetrics contains session metrics enriched with project and experiment context.
type EnrichedMetrics struct {
	SessionID      string
	ProjectID      string
	ProjectName    string
	ExperimentID   *string
	ExperimentName *string

	TokenInput      int64
	TokenOutput     int64
	TokenCacheRead  int64
	TokenCacheWrite int64
	CostEstimateUSD float64

	DurationSeconds int64
	TurnCount       int64
	ErrorCount      int64
	ExitReason      string

	StartedAt time.Time
	EndedAt   time.Time
}
