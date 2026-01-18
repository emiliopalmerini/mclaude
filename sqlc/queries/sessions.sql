-- name: CreateSession :exec
INSERT OR REPLACE INTO sessions (id, project_id, experiment_id, transcript_path, transcript_stored_path, cwd, permission_mode, exit_reason, started_at, ended_at, duration_seconds, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE id = ?;

-- name: ListSessions :many
SELECT * FROM sessions
ORDER BY created_at DESC
LIMIT ?;

-- name: ListSessionsByProject :many
SELECT * FROM sessions
WHERE project_id = ?
ORDER BY created_at DESC
LIMIT ?;

-- name: ListSessionsByExperiment :many
SELECT * FROM sessions
WHERE experiment_id = ?
ORDER BY created_at DESC
LIMIT ?;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = ?;

-- name: DeleteSessionsBefore :execrows
DELETE FROM sessions WHERE created_at < ?;

-- name: DeleteSessionsByProject :execrows
DELETE FROM sessions WHERE project_id = ?;

-- name: DeleteSessionsByExperiment :execrows
DELETE FROM sessions WHERE experiment_id = ?;

-- name: GetSessionTranscriptPaths :many
SELECT id, transcript_stored_path FROM sessions WHERE transcript_stored_path IS NOT NULL;

-- name: GetSessionTranscriptPathsBefore :many
SELECT id, transcript_stored_path FROM sessions WHERE created_at < ? AND transcript_stored_path IS NOT NULL;

-- name: GetSessionTranscriptPathsByProject :many
SELECT id, transcript_stored_path FROM sessions WHERE project_id = ? AND transcript_stored_path IS NOT NULL;

-- name: GetSessionTranscriptPathsByExperiment :many
SELECT id, transcript_stored_path FROM sessions WHERE experiment_id = ? AND transcript_stored_path IS NOT NULL;
