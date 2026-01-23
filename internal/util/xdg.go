package util

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetXDGDataDir returns the XDG data directory for mclaude.
// It respects XDG_DATA_HOME if set, otherwise falls back to ~/.local/share/mclaude
func GetXDGDataDir() (string, error) {
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return filepath.Join(dataHome, "mclaude"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".local", "share", "mclaude"), nil
}
