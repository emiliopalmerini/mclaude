-- name: StoreTranscript :exec
INSERT OR REPLACE INTO session_transcripts (session_id, gzip_data, created_at)
VALUES (?, ?, datetime('now'));

-- name: GetTranscript :one
SELECT gzip_data FROM session_transcripts WHERE session_id = ?;

-- name: DeleteTranscript :exec
DELETE FROM session_transcripts WHERE session_id = ?;

-- name: TranscriptExists :one
SELECT COUNT(*) FROM session_transcripts WHERE session_id = ?;
