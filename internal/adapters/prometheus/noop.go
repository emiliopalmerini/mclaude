package prometheus

import (
	"context"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

// NoOpClient is a Prometheus client that always returns unavailable.
type NoOpClient struct{}

// NewNoOpClient creates a new no-op client for graceful degradation.
func NewNoOpClient() *NoOpClient {
	return &NoOpClient{}
}

func (c *NoOpClient) GetRollingWindowUsage(ctx context.Context, hours int) (*ports.UsageWindow, error) {
	return &ports.UsageWindow{
		TotalTokens: 0,
		TotalCost:   0,
		WindowHours: hours,
		Available:   false,
	}, nil
}

func (c *NoOpClient) IsAvailable(ctx context.Context) bool {
	return false
}
