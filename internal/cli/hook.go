package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/adapters/turso"
	"github.com/emiliopalmerini/mclaude/internal/domain"
)

// HookResponse is the standard output format for hook commands.
// Claude Code reads this from stdout to process hook results.
type HookResponse struct {
	AdditionalContext  string          `json:"additionalContext,omitempty"`
	HookSpecificOutput json.RawMessage `json:"hookSpecificOutput,omitempty"`
	Decision           string          `json:"decision,omitempty"`
	Reason             string          `json:"reason,omitempty"`
}

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Handle Claude Code hook events",
	Long: `Reads hook event JSON from stdin and dispatches to the appropriate handler.

This is a unified entry point for all Claude Code hook events. Configure
your hooks to use "mclaude hook" for all event types:

  {
    "hooks": {
      "SessionStart": [{"type": "command", "command": "mclaude hook"}],
      "SessionEnd":   [{"type": "command", "command": "mclaude hook", "async": true}],
      "Stop":         [{"type": "command", "command": "mclaude hook", "async": true}],
      "PostToolUse":  [{"type": "command", "command": "mclaude hook", "async": true}],
      "SubagentStart":[{"type": "command", "command": "mclaude hook", "async": true}],
      "SubagentStop": [{"type": "command", "command": "mclaude hook", "async": true}]
    }
  }`,
	RunE: runHook,
}

func runHook(cmd *cobra.Command, args []string) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	event, err := domain.ParseHookEvent(input)
	if err != nil {
		return fmt.Errorf("failed to parse hook event: %w", err)
	}

	switch e := event.(type) {
	case *domain.SessionEndInput:
		return handleSessionEnd(e)
	case *domain.SessionStartInput:
		return handleSessionStart(e)
	case *domain.StopInput:
		return handleStop(e)
	case *domain.PostToolUseInput:
		return handlePostToolUse(e)
	case *domain.SubagentStartInput:
		return handleSubagentStart(e)
	case *domain.SubagentStopInput:
		return handleSubagentStop(e)
	default:
		return fmt.Errorf("unhandled hook event type: %T", event)
	}
}

// handleSessionEnd processes a SessionEnd event using the existing record logic.
func handleSessionEnd(event *domain.SessionEndInput) error {
	hookInput := &domain.HookInput{
		SessionID:      event.SessionID,
		TranscriptPath: event.TranscriptPath,
		Cwd:            event.Cwd,
		PermissionMode: event.PermissionMode,
		HookEventName:  event.HookEventName,
		Reason:         event.Reason,
	}
	return processRecordInput(hookInput)
}

// hookDB returns a database connection and cleanup function.
// Uses testDBOverride if set (for tests), otherwise creates a new Turso connection.
func hookDB() (*sql.DB, *turso.DB, func(), error) {
	if testDBOverride != nil {
		return testDBOverride, nil, func() {}, nil
	}

	db, err := turso.NewDB()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db.DB, db, func() { db.Close() }, nil
}

// syncAndClose syncs the Turso DB (if present) and runs the cleanup function.
func syncAndClose(tursoDB *turso.DB, closeDB func()) {
	if tursoDB != nil {
		if err := tursoDB.Sync(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: sync failed: %v\n", err)
		}
	}
	closeDB()
}

// outputJSON writes a HookResponse as JSON to stdout.
func outputJSON(resp *HookResponse) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
