package analytics

import "context"

// Service provides analytics business logic
type Service struct {
	repo   Repository
	logger Logger
}

// NewService creates a new analytics service
func NewService(repo Repository, logger Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// GetOverview returns aggregate metrics for the dashboard overview
func (s *Service) GetOverview(ctx context.Context) (OverviewMetrics, error) {
	s.logger.Debug("Fetching overview metrics")
	return s.repo.GetOverviewMetrics(ctx)
}

// ListSessions returns filtered session summaries with total count
func (s *Service) ListSessions(ctx context.Context, filter SessionFilter) ([]SessionSummary, int, error) {
	s.logger.Debug("Listing sessions")
	return s.repo.ListSessions(ctx, filter)
}

// GetSession returns detailed session information
func (s *Service) GetSession(ctx context.Context, sessionID string) (SessionDetail, error) {
	s.logger.Debug("Getting session detail")
	return s.repo.GetSession(ctx, sessionID)
}

// GetCostsBreakdown returns cost analysis data
func (s *Service) GetCostsBreakdown(ctx context.Context) (CostsBreakdown, error) {
	s.logger.Debug("Getting costs breakdown")
	return s.repo.GetCostsBreakdown(ctx)
}
