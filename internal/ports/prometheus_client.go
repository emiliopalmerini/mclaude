package ports

import (
	"context"
)

// PrometheusClient queries Prometheus for real-time usage metrics.
type PrometheusClient interface {
	// GetRollingWindowUsage retrieves aggregated usage for the specified rolling window.
	GetRollingWindowUsage(ctx context.Context, hours int) (*UsageWindow, error)
	// IsAvailable checks if Prometheus is reachable.
	IsAvailable(ctx context.Context) bool
}

// UsageWindow contains aggregated usage metrics for a time window.
type UsageWindow struct {
	TotalTokens float64
	TotalCost   float64
	WindowHours int
	Available   bool
}
