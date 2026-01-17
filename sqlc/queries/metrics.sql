-- name: CreateSessionMetrics :exec
INSERT INTO session_metrics (session_id, message_count_user, message_count_assistant, turn_count, token_input, token_output, token_cache_read, token_cache_write, cost_estimate_usd, error_count)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetSessionMetricsBySessionID :one
SELECT * FROM session_metrics WHERE session_id = ?;

-- name: CreateSessionTool :exec
INSERT INTO session_tools (session_id, tool_name, invocation_count, total_duration_ms, error_count)
VALUES (?, ?, ?, ?, ?)
ON CONFLICT (session_id, tool_name) DO UPDATE SET
    invocation_count = invocation_count + excluded.invocation_count,
    error_count = error_count + excluded.error_count;

-- name: ListSessionToolsBySessionID :many
SELECT * FROM session_tools WHERE session_id = ? ORDER BY invocation_count DESC;

-- name: CreateSessionFile :exec
INSERT INTO session_files (session_id, file_path, operation, operation_count)
VALUES (?, ?, ?, ?)
ON CONFLICT (session_id, file_path, operation) DO UPDATE SET
    operation_count = operation_count + excluded.operation_count;

-- name: ListSessionFilesBySessionID :many
SELECT * FROM session_files WHERE session_id = ? ORDER BY operation_count DESC;

-- name: CreateSessionCommand :exec
INSERT INTO session_commands (session_id, command, exit_code, executed_at)
VALUES (?, ?, ?, ?);

-- name: ListSessionCommandsBySessionID :many
SELECT * FROM session_commands WHERE session_id = ? ORDER BY id ASC;

-- name: GetAggregateStats :one
SELECT
    COUNT(DISTINCT s.id) as session_count,
    COALESCE(SUM(m.message_count_user), 0) as total_user_messages,
    COALESCE(SUM(m.message_count_assistant), 0) as total_assistant_messages,
    COALESCE(SUM(m.turn_count), 0) as total_turns,
    COALESCE(SUM(m.token_input), 0) as total_token_input,
    COALESCE(SUM(m.token_output), 0) as total_token_output,
    COALESCE(SUM(m.token_cache_read), 0) as total_token_cache_read,
    COALESCE(SUM(m.token_cache_write), 0) as total_token_cache_write,
    COALESCE(SUM(m.cost_estimate_usd), 0) as total_cost_usd,
    COALESCE(SUM(m.error_count), 0) as total_errors
FROM sessions s
LEFT JOIN session_metrics m ON s.id = m.session_id
WHERE s.created_at >= ?;

-- name: GetAggregateStatsByExperiment :one
SELECT
    COUNT(DISTINCT s.id) as session_count,
    COALESCE(SUM(m.message_count_user), 0) as total_user_messages,
    COALESCE(SUM(m.message_count_assistant), 0) as total_assistant_messages,
    COALESCE(SUM(m.turn_count), 0) as total_turns,
    COALESCE(SUM(m.token_input), 0) as total_token_input,
    COALESCE(SUM(m.token_output), 0) as total_token_output,
    COALESCE(SUM(m.token_cache_read), 0) as total_token_cache_read,
    COALESCE(SUM(m.token_cache_write), 0) as total_token_cache_write,
    COALESCE(SUM(m.cost_estimate_usd), 0) as total_cost_usd,
    COALESCE(SUM(m.error_count), 0) as total_errors
FROM sessions s
LEFT JOIN session_metrics m ON s.id = m.session_id
WHERE s.experiment_id = ? AND s.created_at >= ?;

-- name: GetAggregateStatsByProject :one
SELECT
    COUNT(DISTINCT s.id) as session_count,
    COALESCE(SUM(m.message_count_user), 0) as total_user_messages,
    COALESCE(SUM(m.message_count_assistant), 0) as total_assistant_messages,
    COALESCE(SUM(m.turn_count), 0) as total_turns,
    COALESCE(SUM(m.token_input), 0) as total_token_input,
    COALESCE(SUM(m.token_output), 0) as total_token_output,
    COALESCE(SUM(m.token_cache_read), 0) as total_token_cache_read,
    COALESCE(SUM(m.token_cache_write), 0) as total_token_cache_write,
    COALESCE(SUM(m.cost_estimate_usd), 0) as total_cost_usd,
    COALESCE(SUM(m.error_count), 0) as total_errors
FROM sessions s
LEFT JOIN session_metrics m ON s.id = m.session_id
WHERE s.project_id = ? AND s.created_at >= ?;

-- name: GetTopToolsUsage :many
SELECT
    tool_name,
    SUM(invocation_count) as total_invocations,
    SUM(error_count) as total_errors
FROM session_tools st
JOIN sessions s ON st.session_id = s.id
WHERE s.created_at >= ?
GROUP BY tool_name
ORDER BY total_invocations DESC
LIMIT ?;
