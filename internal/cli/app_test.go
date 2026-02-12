package cli

import (
	"testing"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

func TestAppContextFieldTypes(t *testing.T) {
	// Compile-time verification that AppContext uses port interfaces.
	var a AppContext
	var _ ports.SessionRepository = a.SessionRepo                      //nolint:staticcheck
	var _ ports.SessionMetricsRepository = a.MetricsRepo               //nolint:staticcheck
	var _ ports.SessionToolRepository = a.ToolRepo                     //nolint:staticcheck
	var _ ports.SessionFileRepository = a.FileRepo                     //nolint:staticcheck
	var _ ports.SessionCommandRepository = a.CommandRepo               //nolint:staticcheck
	var _ ports.SessionSubagentRepository = a.SubagentRepo             //nolint:staticcheck
	var _ ports.ExperimentRepository = a.ExperimentRepo                //nolint:staticcheck
	var _ ports.ExperimentVariableRepository = a.ExpVariableRepo       //nolint:staticcheck
	var _ ports.ProjectRepository = a.ProjectRepo                      //nolint:staticcheck
	var _ ports.PricingRepository = a.PricingRepo                      //nolint:staticcheck
	var _ ports.SessionQualityRepository = a.QualityRepo               //nolint:staticcheck
	var _ ports.PlanConfigRepository = a.PlanConfigRepo                //nolint:staticcheck
	var _ ports.StatsRepository = a.StatsRepo                          //nolint:staticcheck
	var _ ports.TranscriptStorage = a.TranscriptStorage                //nolint:staticcheck
}

func TestAppContextClose_NilDB(t *testing.T) {
	a := &AppContext{}
	if err := a.Close(); err != nil {
		t.Errorf("Close() on nil DB should not error, got: %v", err)
	}
}
