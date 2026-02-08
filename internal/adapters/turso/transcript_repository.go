package turso

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"

	sqlc "github.com/emiliopalmerini/mclaude/sqlc/generated"
)

type TranscriptRepository struct {
	queries *sqlc.Queries
}

func NewTranscriptRepository(db *sql.DB) *TranscriptRepository {
	return &TranscriptRepository{
		queries: sqlc.New(db),
	}
}

func (r *TranscriptRepository) Store(ctx context.Context, sessionID string, sourcePath string) (string, error) {
	src, err := os.Open(sourcePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = src.Close() }()

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)

	if _, err := io.Copy(gw, src); err != nil {
		return "", fmt.Errorf("failed to compress transcript: %w", err)
	}
	if err := gw.Close(); err != nil {
		return "", fmt.Errorf("failed to close gzip writer: %w", err)
	}

	if err := r.queries.StoreTranscript(ctx, sqlc.StoreTranscriptParams{
		SessionID: sessionID,
		GzipData:  buf.Bytes(),
	}); err != nil {
		return "", fmt.Errorf("failed to store transcript in database: %w", err)
	}

	return "db", nil
}

func (r *TranscriptRepository) Get(ctx context.Context, sessionID string) ([]byte, error) {
	gzipData, err := r.queries.GetTranscript(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transcript from database: %w", err)
	}

	gr, err := gzip.NewReader(bytes.NewReader(gzipData))
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() { _ = gr.Close() }()

	data, err := io.ReadAll(gr)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress transcript: %w", err)
	}

	return data, nil
}

func (r *TranscriptRepository) Delete(ctx context.Context, sessionID string) error {
	return r.queries.DeleteTranscript(ctx, sessionID)
}

func (r *TranscriptRepository) Exists(ctx context.Context, sessionID string) (bool, error) {
	count, err := r.queries.TranscriptExists(ctx, sessionID)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
