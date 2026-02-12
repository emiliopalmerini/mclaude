package cli

import (
	"fmt"

	"github.com/emiliopalmerini/mclaude/internal/adapters/storage"
	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/ports"
)

// AppContext holds all shared dependencies for CLI commands.
type AppContext struct {
	DB               *turso.DB
	SessionRepo      ports.SessionRepository
	MetricsRepo      ports.SessionMetricsRepository
	ToolRepo         ports.SessionToolRepository
	FileRepo         ports.SessionFileRepository
	CommandRepo      ports.SessionCommandRepository
	SubagentRepo     ports.SessionSubagentRepository
	ExperimentRepo    ports.ExperimentRepository
	ExpVariableRepo   ports.ExperimentVariableRepository
	ProjectRepo       ports.ProjectRepository
	PricingRepo      ports.PricingRepository
	QualityRepo      ports.SessionQualityRepository
	PlanConfigRepo    ports.PlanConfigRepository
	StatsRepo         ports.StatsRepository
	TranscriptStorage ports.TranscriptStorage
}

// NewAppContext creates an AppContext with all dependencies initialized.
func NewAppContext() (*AppContext, error) {
	db, err := turso.NewDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	transcriptStorage, err := storage.NewTranscriptStorage()
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize transcript storage: %w", err)
	}

	return &AppContext{
		DB:                db,
		SessionRepo:       turso.NewSessionRepository(db.DB),
		MetricsRepo:       turso.NewSessionMetricsRepository(db.DB),
		ToolRepo:          turso.NewSessionToolRepository(db.DB),
		FileRepo:          turso.NewSessionFileRepository(db.DB),
		CommandRepo:       turso.NewSessionCommandRepository(db.DB),
		SubagentRepo:      turso.NewSessionSubagentRepository(db.DB),
		ExperimentRepo:    turso.NewExperimentRepository(db.DB),
		ExpVariableRepo:   turso.NewExperimentVariableRepository(db.DB),
		ProjectRepo:       turso.NewProjectRepository(db.DB),
		PricingRepo:       turso.NewPricingRepository(db.DB),
		QualityRepo:       turso.NewSessionQualityRepository(db.DB),
		PlanConfigRepo:    turso.NewPlanConfigRepository(db.DB),
		StatsRepo:         turso.NewStatsRepository(db.DB),
		TranscriptStorage: transcriptStorage,
	}, nil
}

// Close releases all resources held by the AppContext.
func (a *AppContext) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
