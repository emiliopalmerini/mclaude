-- name: GetDashboardMetrics :one
SELECT
    COUNT(*) as total_sessions,
    COALESCE(SUM(estimated_cost_usd), 0) as total_cost_usd,
    COALESCE(SUM(input_tokens), 0) as total_input_tokens,
    COALESCE(SUM(output_tokens), 0) as total_output_tokens,
    COALESCE(SUM(thinking_tokens), 0) as total_thinking_tokens,
    COALESCE(SUM(cache_read_tokens), 0) as total_cache_read_tokens,
    COALESCE(SUM(cache_write_tokens), 0) as total_cache_write_tokens,
    COALESCE(SUM(tool_calls), 0) as total_tool_calls,
    COALESCE(AVG(duration_seconds), 0) as avg_duration_seconds
FROM sessions;

-- name: GetTodayMetrics :one
SELECT
    COUNT(*) as sessions_today,
    COALESCE(SUM(estimated_cost_usd), 0) as cost_today
FROM sessions
WHERE date(timestamp) = date('now');

-- name: GetWeekMetrics :one
SELECT
    COUNT(*) as sessions_week,
    COALESCE(SUM(estimated_cost_usd), 0) as cost_week
FROM sessions
WHERE date(timestamp) >= date('now', '-7 days');

-- name: ListSessions :many
SELECT
    id, session_id, hostname, timestamp, exit_reason,
    working_directory, git_branch, duration_seconds,
    user_prompts, tool_calls, estimated_cost_usd, model
FROM sessions
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: CountSessions :one
SELECT COUNT(*) as count FROM sessions;

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE session_id = ?;

-- name: GetDistinctHostnames :many
SELECT DISTINCT hostname FROM sessions WHERE hostname != '' ORDER BY hostname;

-- name: GetDistinctBranches :many
SELECT DISTINCT git_branch FROM sessions WHERE git_branch != '' ORDER BY git_branch;

-- name: GetDistinctModels :many
SELECT DISTINCT model FROM sessions WHERE model != '' ORDER BY model;
