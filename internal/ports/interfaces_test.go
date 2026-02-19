package ports_test

import (
	"testing"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/ports"
)

// Compile-time interface conformance checks.
// These verify that concrete adapters properly implement their port interfaces.

func TestSessionRepositoryConformance(t *testing.T) {
	var _ ports.SessionRepository = (*turso.SessionRepository)(nil)
}

func TestSessionMetricsRepositoryConformance(t *testing.T) {
	var _ ports.SessionMetricsRepository = (*turso.SessionMetricsRepository)(nil)
}

func TestSessionToolRepositoryConformance(t *testing.T) {
	var _ ports.SessionToolRepository = (*turso.SessionToolRepository)(nil)
}

func TestSessionFileRepositoryConformance(t *testing.T) {
	var _ ports.SessionFileRepository = (*turso.SessionFileRepository)(nil)
}

func TestSessionCommandRepositoryConformance(t *testing.T) {
	var _ ports.SessionCommandRepository = (*turso.SessionCommandRepository)(nil)
}

func TestSessionSubagentRepositoryConformance(t *testing.T) {
	var _ ports.SessionSubagentRepository = (*turso.SessionSubagentRepository)(nil)
}

func TestExperimentRepositoryConformance(t *testing.T) {
	var _ ports.ExperimentRepository = (*turso.ExperimentRepository)(nil)
}

func TestExperimentVariableRepositoryConformance(t *testing.T) {
	var _ ports.ExperimentVariableRepository = (*turso.ExperimentVariableRepository)(nil)
}

func TestProjectRepositoryConformance(t *testing.T) {
	var _ ports.ProjectRepository = (*turso.ProjectRepository)(nil)
}

func TestPricingRepositoryConformance(t *testing.T) {
	var _ ports.PricingRepository = (*turso.PricingRepository)(nil)
}

func TestStatsRepositoryConformance(t *testing.T) {
	var _ ports.StatsRepository = (*turso.StatsRepository)(nil)
}

func TestToolEventRepositoryConformance(t *testing.T) {
	var _ ports.ToolEventRepository = (*turso.ToolEventRepository)(nil)
}
