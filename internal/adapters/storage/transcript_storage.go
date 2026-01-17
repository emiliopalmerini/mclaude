package storage

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type TranscriptStorage struct {
	baseDir string
}

func NewTranscriptStorage() (*TranscriptStorage, error) {
	baseDir, err := getXDGDataDir()
	if err != nil {
		return nil, err
	}

	transcriptsDir := filepath.Join(baseDir, "transcripts")
	if err := os.MkdirAll(transcriptsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create transcripts directory: %w", err)
	}

	return &TranscriptStorage{baseDir: transcriptsDir}, nil
}

func (s *TranscriptStorage) Store(ctx context.Context, sessionID string, sourcePath string) (string, error) {
	destPath := s.getPath(sessionID)

	// Open source file
	src, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dest, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dest.Close()

	// Create gzip writer
	gw := gzip.NewWriter(dest)
	defer gw.Close()

	// Copy with gzip compression
	if _, err := io.Copy(gw, src); err != nil {
		return "", fmt.Errorf("failed to compress transcript: %w", err)
	}

	// Ensure gzip writer is flushed
	if err := gw.Close(); err != nil {
		return "", fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return destPath, nil
}

func (s *TranscriptStorage) Get(ctx context.Context, sessionID string) ([]byte, error) {
	path := s.getPath(sessionID)

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open transcript file: %w", err)
	}
	defer file.Close()

	gr, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gr.Close()

	data, err := io.ReadAll(gr)
	if err != nil {
		return nil, fmt.Errorf("failed to read transcript: %w", err)
	}

	return data, nil
}

func (s *TranscriptStorage) Delete(ctx context.Context, sessionID string) error {
	path := s.getPath(sessionID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete transcript: %w", err)
	}
	return nil
}

func (s *TranscriptStorage) Exists(ctx context.Context, sessionID string) (bool, error) {
	path := s.getPath(sessionID)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *TranscriptStorage) getPath(sessionID string) string {
	return filepath.Join(s.baseDir, sessionID+".jsonl.gz")
}

func getXDGDataDir() (string, error) {
	// Check XDG_DATA_HOME first
	if dataHome := os.Getenv("XDG_DATA_HOME"); dataHome != "" {
		return filepath.Join(dataHome, "claude-watcher"), nil
	}

	// Fall back to ~/.local/share
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, ".local", "share", "claude-watcher"), nil
}
