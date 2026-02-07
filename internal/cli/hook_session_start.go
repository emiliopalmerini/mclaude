package cli

import (
	"context"
	"fmt"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/domain"
)

func handleSessionStart(event *domain.SessionStartInput) error {
	sqlDB, _, closeDB, err := hookDB()
	if err != nil {
		return err
	}
	defer closeDB()

	ctx := context.Background()
	experimentRepo := turso.NewExperimentRepository(sqlDB)

	activeExperiment, err := experimentRepo.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active experiment: %w", err)
	}

	// No active experiment → no output needed
	if activeExperiment == nil {
		return nil
	}

	// Build context string with experiment info
	contextStr := fmt.Sprintf("Active experiment: %s", activeExperiment.Name)
	if activeExperiment.Hypothesis != nil {
		contextStr += fmt.Sprintf("\nHypothesis: %s", *activeExperiment.Hypothesis)
	}
	if activeExperiment.Description != nil {
		contextStr += fmt.Sprintf("\nDescription: %s", *activeExperiment.Description)
	}

	return outputJSON(&HookResponse{
		AdditionalContext: contextStr,
	})
}
