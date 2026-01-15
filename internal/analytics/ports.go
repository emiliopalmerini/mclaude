package analytics

import "context"

// Repository defines the interface for analytics data access
type Repository interface {
	// GetOverviewMetrics retrieves aggregate metrics for the dashboard overview
	GetOverviewMetrics(ctx context.Context) (OverviewMetrics, error)

	// ListSessions returns paginated session summaries with total count
	ListSessions(ctx context.Context, filter SessionFilter) ([]SessionSummary, int, error)

	// GetSession retrieves detailed session information
	GetSession(ctx context.Context, sessionID string) (SessionDetail, error)

	// GetCostsBreakdown retrieves cost analysis data
	GetCostsBreakdown(ctx context.Context) (CostsBreakdown, error)
}

// Logger defines the interface for logging
type Logger interface {
	Debug(msg string)
	Error(msg string)
}
