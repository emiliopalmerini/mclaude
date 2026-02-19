package turso

import (
	"database/sql"

	"github.com/emiliopalmerini/mclaude/internal/ports"
)

// Repositories holds all turso repository implementations as port interfaces.
type Repositories struct {
	Sessions            ports.SessionRepository
	Metrics             ports.SessionMetricsRepository
	Tools               ports.SessionToolRepository
	Files               ports.SessionFileRepository
	Commands            ports.SessionCommandRepository
	Subagents           ports.SessionSubagentRepository
	ToolEvents          ports.ToolEventRepository
	Experiments         ports.ExperimentRepository
	ExperimentVariables ports.ExperimentVariableRepository
	Projects            ports.ProjectRepository
	Pricing             ports.PricingRepository
	Stats               ports.StatsRepository
}

// NewRepositories creates all turso repository implementations from a database connection.
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Sessions:            NewSessionRepository(db),
		Metrics:             NewSessionMetricsRepository(db),
		Tools:               NewSessionToolRepository(db),
		Files:               NewSessionFileRepository(db),
		Commands:            NewSessionCommandRepository(db),
		Subagents:           NewSessionSubagentRepository(db),
		ToolEvents:          NewToolEventRepository(db),
		Experiments:         NewExperimentRepository(db),
		ExperimentVariables: NewExperimentVariableRepository(db),
		Projects:            NewProjectRepository(db),
		Pricing:             NewPricingRepository(db),
		Stats:               NewStatsRepository(db),
	}
}
