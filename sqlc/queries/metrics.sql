-- name: CreateSessionMetrics :exec
INSERT OR REPLACE INTO session_metrics (session_id, model_id, message_count_user, message_count_assistant, turn_count, token_input, token_output, token_cache_read, token_cache_write, cost_estimate_usd, error_count, input_rate, output_rate, cache_read_rate, cache_write_rate)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

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

-- name: GetDailyStats :many
SELECT
    DATE(s.created_at) as date,
    COUNT(DISTINCT s.id) as session_count,
    COALESCE(SUM(m.token_input + m.token_output), 0) as total_tokens,
    COALESCE(SUM(m.cost_estimate_usd), 0) as total_cost
FROM sessions s
LEFT JOIN session_metrics m ON s.id = m.session_id
WHERE s.created_at >= ?
GROUP BY DATE(s.created_at)
ORDER BY date ASC
LIMIT ?;

-- name: GetStatsForAllExperiments :many
SELECT
    e.id as experiment_id,
    e.name as experiment_name,
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
FROM experiments e
LEFT JOIN sessions s ON s.experiment_id = e.id
LEFT JOIN session_metrics m ON s.id = m.session_id
GROUP BY e.id, e.name
ORDER BY e.created_at DESC;

-- name: CreateSessionSubagent :exec
INSERT INTO session_subagents (session_id, agent_type, agent_kind, description, model, total_tokens, token_input, token_output, token_cache_read, token_cache_write, total_duration_ms, tool_use_count, cost_estimate_usd)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: ListSessionSubagentsBySessionID :many
SELECT * FROM session_subagents WHERE session_id = ? ORDER BY id ASC;

-- name: GetSubagentStatsBySession :many
SELECT
    agent_type,
    agent_kind,
    COUNT(*) as invocation_count,
    COALESCE(SUM(total_tokens), 0) as total_tokens,
    COALESCE(SUM(cost_estimate_usd), 0) as total_cost
FROM session_subagents
WHERE session_id = ?
GROUP BY agent_type, agent_kind
ORDER BY total_tokens DESC;

-- name: GetTopSubagentUsage :many
SELECT
    agent_type,
    agent_kind,
    COUNT(*) as invocation_count,
    COALESCE(SUM(total_tokens), 0) as total_tokens,
    COALESCE(SUM(cost_estimate_usd), 0) as total_cost
FROM session_subagents sa
JOIN sessions s ON sa.session_id = s.id
WHERE s.created_at >= ?
GROUP BY agent_type, agent_kind
ORDER BY total_tokens DESC
LIMIT ?;

-- name: GetTotalToolCallsByExperiment :one
SELECT COALESCE(SUM(st.invocation_count), 0) as total_tool_calls
FROM session_tools st
JOIN sessions s ON st.session_id = s.id
WHERE s.experiment_id = ?;

-- name: GetTopToolsUsageByExperiment :many
SELECT
    tool_name,
    SUM(invocation_count) as total_invocations,
    SUM(error_count) as total_errors
FROM session_tools st
JOIN sessions s ON st.session_id = s.id
WHERE s.experiment_id = ?
GROUP BY tool_name
ORDER BY total_invocations DESC
LIMIT ?;
