package cli

import (
	"context"

	"github.com/emiliopalmerini/mclaude/internal/domain"
)

func handleStop(event *domain.StopInput) error {
	// Prevent infinite loop: if this stop hook triggered another stop, bail out
	if event.StopHookActive {
		return nil
	}

	sqlDB, tursoDB, closeDB, err := hookDB()
	if err != nil {
		return err
	}
	defer func() { syncAndClose(tursoDB, closeDB) }()

	return saveSessionData(context.Background(), sqlDB, event.SessionID, event.TranscriptPath, event.Cwd, event.PermissionMode, saveSessionOpts{
		SkipTranscriptStorage: true, // transcript file is still being written
	})
}
