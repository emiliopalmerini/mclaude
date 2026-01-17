package ports

import "context"

type TranscriptStorage interface {
	Store(ctx context.Context, sessionID string, sourcePath string) (storedPath string, err error)
	Get(ctx context.Context, sessionID string) ([]byte, error)
	Delete(ctx context.Context, sessionID string) error
	Exists(ctx context.Context, sessionID string) (bool, error)
}
