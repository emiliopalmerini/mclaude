package cli

import (
	"context"
	"fmt"

	"github.com/emiliopalmerini/mclaude/internal/domain"
	"github.com/emiliopalmerini/mclaude/internal/ports"
	"github.com/emiliopalmerini/mclaude/internal/util"
)

// getExperimentByName looks up an experiment by name via the repository.
// Returns a descriptive error if not found or if the lookup fails.
func getExperimentByName(ctx context.Context, repo ports.ExperimentRepository, name string) (*domain.Experiment, error) {
	exp, err := repo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiment: %w", err)
	}
	if exp == nil {
		return nil, fmt.Errorf("experiment %q not found", name)
	}
	return exp, nil
}

// truncate shortens a string to maxLen characters, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func formatDateTimeCLI(s string) string {
	return util.FormatDateTime(s)
}

func formatTokensCLI(n int64) string {
	return util.FormatNumber(n)
}

func shortModel(modelID string) string {
	s := modelID
	if len(s) > 7 && s[:7] == "claude-" {
		s = s[7:]
	}
	// Strip trailing date suffix (e.g. "-20250929")
	if len(s) > 9 && s[len(s)-9] == '-' {
		candidate := s[len(s)-8:]
		allDigits := true
		for _, c := range candidate {
			if c < '0' || c > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			s = s[:len(s)-9]
		}
	}
	return s
}

func formatDurationCLI(seconds int64) string {
	if seconds == 0 {
		return "-"
	}
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	return fmt.Sprintf("%dm%ds", seconds/60, seconds%60)
}
