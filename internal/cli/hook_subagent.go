package cli

import (
	"fmt"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

func handleSubagentStart(event *domain.SubagentStartInput) error {
	return fmt.Errorf("SubagentStart handler not yet implemented")
}

func handleSubagentStop(event *domain.SubagentStopInput) error {
	return fmt.Errorf("SubagentStop handler not yet implemented")
}
