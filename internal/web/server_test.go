package web

import (
	"testing"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

func TestServerFieldsAreInterfaces(t *testing.T) {
	s := &Server{}
	var _ ports.TranscriptStorage = s.transcriptStorage
	var _ ports.SessionQualityRepository = s.qualityRepo
	var _ ports.PlanConfigRepository = s.planConfigRepo
	var _ ports.PrometheusClient = s.promClient
}
