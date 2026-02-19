package web

import (
	"testing"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

func TestServerFieldsAreInterfaces(t *testing.T) {
	s := &Server{}
	var _ ports.ExperimentRepository = s.experimentRepo          //nolint:staticcheck
	var _ ports.ExperimentVariableRepository = s.expVariableRepo //nolint:staticcheck
	var _ ports.PricingRepository = s.pricingRepo                //nolint:staticcheck
	var _ ports.SessionRepository = s.sessionRepo                //nolint:staticcheck
	var _ ports.SessionMetricsRepository = s.metricsRepo         //nolint:staticcheck
	var _ ports.StatsRepository = s.statsRepo                    //nolint:staticcheck
	var _ ports.ProjectRepository = s.projectRepo                //nolint:staticcheck
}
