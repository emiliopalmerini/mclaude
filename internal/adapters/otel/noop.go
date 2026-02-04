package otel

import (
	"context"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

// NoOpExporter is a metrics exporter that does nothing.
type NoOpExporter struct{}

// NewNoOpExporter creates a new no-op exporter for graceful degradation.
func NewNoOpExporter() *NoOpExporter {
	return &NoOpExporter{}
}

func (e *NoOpExporter) ExportSessionMetrics(ctx context.Context, m *ports.EnrichedMetrics) error {
	return nil
}

func (e *NoOpExporter) Close(ctx context.Context) error {
	return nil
}
