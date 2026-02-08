package cli

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

// testDBOverride allows tests to inject a database connection.
// When set, hookDB uses this instead of creating a new connection.
var testDBOverride *sql.DB

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record session data from Claude Code hook",
	Long: `Reads session data from stdin (Claude Code SessionEnd hook),
parses the transcript, and saves all data to the database.

This command is designed to be called from a Claude Code hook with
"async": true so it runs in the background:

  {
    "hooks": {
      "SessionEnd": [
        {
          "hooks": [
            {
              "type": "command",
              "command": "mclaude record",
              "async": true
            }
          ]
        }
      ]
    }
  }`,
	RunE: runRecord,
}

func runRecord(cmd *cobra.Command, args []string) error {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %w", err)
	}

	var hookInput domain.HookInput
	if err := json.Unmarshal(input, &hookInput); err != nil {
		return fmt.Errorf("failed to parse hook input: %w", err)
	}

	return processRecordInput(&hookInput)
}

func processRecordInput(hookInput *domain.HookInput) error {
	sqlDB, tursoDB, closeDB, err := hookDB()
	if err != nil {
		return err
	}
	defer func() { syncAndClose(tursoDB, closeDB) }()

	return saveSessionData(context.Background(), sqlDB, hookInput.SessionID, hookInput.TranscriptPath, hookInput.Cwd, hookInput.PermissionMode, saveSessionOpts{
		ExitReason: hookInput.Reason,
	})
}

// resolveModelAlias maps short model aliases to full model IDs for pricing lookup.
func resolveModelAlias(alias string) string {
	aliases := map[string]string{
		"haiku":  "claude-haiku-4-5-20251001",
		"sonnet": "claude-sonnet-4-5-20250929",
		"opus":   "claude-opus-4-6-20260206",
	}
	if id, ok := aliases[alias]; ok {
		return id
	}
	return alias
}
