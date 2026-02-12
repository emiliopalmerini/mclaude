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

-- name: ListSessionsWithMetrics :many
SELECT
    s.id, s.project_id, s.experiment_id, s.cwd, s.permission_mode, s.exit_reason, s.created_at,
    s.started_at, s.ended_at, s.duration_seconds,
    COALESCE(m.turn_count, 0) as turn_count,
    COALESCE(m.token_input, 0) + COALESCE(m.token_output, 0) as total_tokens,
    m.cost_estimate_usd
FROM sessions s
LEFT JOIN session_metrics m ON s.id = m.session_id
ORDER BY s.created_at DESC
LIMIT ?;

-- name: ListSessionsWithMetricsByExperiment :many
SELECT
    s.id, s.project_id, s.experiment_id, s.cwd, s.permission_mode, s.exit_reason, s.created_at,
    s.started_at, s.ended_at, s.duration_seconds,
    COALESCE(m.turn_count, 0) as turn_count,
    COALESCE(m.token_input, 0) + COALESCE(m.token_output, 0) as total_tokens,
    m.cost_estimate_usd
FROM sessions s
LEFT JOIN session_metrics m ON s.id = m.session_id
WHERE s.experiment_id = ?
ORDER BY s.created_at DESC
LIMIT ?;

-- name: ListSessionsWithMetricsFull :many
SELECT
    s.id, s.project_id, s.experiment_id, s.exit_reason, s.created_at,
    s.duration_seconds,
    COALESCE(m.turn_count, 0) as turn_count,
    COALESCE(m.token_input, 0) + COALESCE(m.token_output, 0) as total_tokens,
    m.cost_estimate_usd,
    m.model_id,
    (SELECT COUNT(*) FROM session_subagents sa WHERE sa.session_id = s.id) as subagent_count
FROM sessions s
LEFT JOIN session_metrics m ON s.id = m.session_id
ORDER BY s.created_at DESC
LIMIT ?;

-- name: ListSessionsWithMetricsFullByExperiment :many
SELECT
    s.id, s.project_id, s.experiment_id, s.exit_reason, s.created_at,
    s.duration_seconds,
    COALESCE(m.turn_count, 0) as turn_count,
    COALESCE(m.token_input, 0) + COALESCE(m.token_output, 0) as total_tokens,
    m.cost_estimate_usd,
    m.model_id,
    (SELECT COUNT(*) FROM session_subagents sa WHERE sa.session_id = s.id) as subagent_count
FROM sessions s
LEFT JOIN session_metrics m ON s.id = m.session_id
WHERE s.experiment_id = ?
ORDER BY s.created_at DESC
LIMIT ?;

-- name: ListSessionsWithMetricsFullByProject :many
SELECT
    s.id, s.project_id, s.experiment_id, s.exit_reason, s.created_at,
    s.duration_seconds,
    COALESCE(m.turn_count, 0) as turn_count,
    COALESCE(m.token_input, 0) + COALESCE(m.token_output, 0) as total_tokens,
    m.cost_estimate_usd,
    m.model_id,
    (SELECT COUNT(*) FROM session_subagents sa WHERE sa.session_id = s.id) as subagent_count
FROM sessions s
LEFT JOIN session_metrics m ON s.id = m.session_id
WHERE s.project_id = ?
ORDER BY s.created_at DESC
LIMIT ?;

-- name: ListSessionsWithMetricsFullByProjectAndExperiment :many
SELECT
    s.id, s.project_id, s.experiment_id, s.exit_reason, s.created_at,
    s.duration_seconds,
    COALESCE(m.turn_count, 0) as turn_count,
    COALESCE(m.token_input, 0) + COALESCE(m.token_output, 0) as total_tokens,
    m.cost_estimate_usd,
    m.model_id,
    (SELECT COUNT(*) FROM session_subagents sa WHERE sa.session_id = s.id) as subagent_count
FROM sessions s
LEFT JOIN session_metrics m ON s.id = m.session_id
WHERE s.project_id = ? AND s.experiment_id = ?
ORDER BY s.created_at DESC
LIMIT ?;
