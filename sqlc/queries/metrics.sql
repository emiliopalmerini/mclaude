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
