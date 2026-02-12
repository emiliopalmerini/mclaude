package web

import (
	"testing"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

func TestServerFieldsAreInterfaces(t *testing.T) {
	s := &Server{}
	var _ ports.TranscriptStorage = s.transcriptStorage          //nolint:staticcheck
	var _ ports.SessionQualityRepository = s.qualityRepo         //nolint:staticcheck
	var _ ports.PlanConfigRepository = s.planConfigRepo          //nolint:staticcheck
}
