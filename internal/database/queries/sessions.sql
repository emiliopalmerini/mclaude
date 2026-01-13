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

-- name: ListSessionsFiltered :many
SELECT
    id, session_id, hostname, timestamp, exit_reason,
    working_directory, git_branch, duration_seconds,
    user_prompts, tool_calls, estimated_cost_usd, model
FROM sessions
WHERE
    (sqlc.narg(hostname) IS NULL OR hostname = sqlc.narg(hostname))
    AND (sqlc.narg(git_branch) IS NULL OR git_branch = sqlc.narg(git_branch))
    AND (sqlc.narg(model) IS NULL OR model = sqlc.narg(model))
    AND (sqlc.narg(start_date) IS NULL OR timestamp >= sqlc.narg(start_date))
    AND (sqlc.narg(end_date) IS NULL OR timestamp <= sqlc.narg(end_date))
ORDER BY timestamp DESC
LIMIT sqlc.arg(limit) OFFSET sqlc.arg(offset);

-- name: CountSessionsFiltered :one
SELECT COUNT(*) as count FROM sessions
WHERE
    (sqlc.narg(hostname) IS NULL OR hostname = sqlc.narg(hostname))
    AND (sqlc.narg(git_branch) IS NULL OR git_branch = sqlc.narg(git_branch))
    AND (sqlc.narg(model) IS NULL OR model = sqlc.narg(model))
    AND (sqlc.narg(start_date) IS NULL OR timestamp >= sqlc.narg(start_date))
    AND (sqlc.narg(end_date) IS NULL OR timestamp <= sqlc.narg(end_date));

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE session_id = ?;

-- name: GetDistinctHostnames :many
SELECT DISTINCT hostname FROM sessions WHERE hostname != '' ORDER BY hostname;

-- name: GetDistinctBranches :many
SELECT DISTINCT git_branch FROM sessions WHERE git_branch != '' ORDER BY git_branch;

-- name: GetDistinctModels :many
SELECT DISTINCT model FROM sessions WHERE model != '' ORDER BY model;

-- name: GetHourlyMetrics :many
SELECT
    strftime('%Y-%m-%dT%H:00:00Z', timestamp) as period,
    COUNT(*) as sessions,
    COALESCE(SUM(estimated_cost_usd), 0) as cost,
    COALESCE(SUM(input_tokens), 0) as input_tokens,
    COALESCE(SUM(output_tokens), 0) as output_tokens,
    COALESCE(SUM(thinking_tokens), 0) as thinking_tokens,
    COALESCE(SUM(cache_read_tokens), 0) as cache_read_tokens,
    COALESCE(SUM(tool_calls), 0) as tool_calls
FROM sessions
WHERE timestamp >= datetime('now', ? || ' hours')
GROUP BY strftime('%Y-%m-%dT%H:00:00Z', timestamp)
ORDER BY period ASC;

-- name: GetDailyMetrics :many
SELECT
    date(timestamp) as period,
    COUNT(*) as sessions,
    COALESCE(SUM(estimated_cost_usd), 0) as cost,
    COALESCE(SUM(input_tokens), 0) as input_tokens,
    COALESCE(SUM(output_tokens), 0) as output_tokens,
    COALESCE(SUM(thinking_tokens), 0) as thinking_tokens,
    COALESCE(SUM(cache_read_tokens), 0) as cache_read_tokens,
    COALESCE(SUM(tool_calls), 0) as tool_calls
FROM sessions
WHERE date(timestamp) >= date('now', ? || ' days')
GROUP BY date(timestamp)
ORDER BY period ASC;

-- name: GetModelDistribution :many
SELECT
    COALESCE(model, 'Unknown') as model,
    COUNT(*) as sessions,
    COALESCE(SUM(estimated_cost_usd), 0) as cost
FROM sessions
WHERE timestamp >= datetime('now', ? || ' hours')
GROUP BY model
ORDER BY sessions DESC;

-- name: GetHourOfDayDistribution :many
SELECT
    CAST(strftime('%H', timestamp) AS INTEGER) as hour,
    COUNT(*) as sessions,
    COALESCE(SUM(estimated_cost_usd), 0) as cost
FROM sessions
WHERE timestamp >= datetime('now', ? || ' hours')
GROUP BY strftime('%H', timestamp)
ORDER BY hour ASC;

-- ============================================================================
-- PRODUCTIVITY & COST ANALYTICS QUERIES
-- ============================================================================

-- name: GetToolsBreakdownAll :many
SELECT tools_breakdown
FROM sessions
WHERE tools_breakdown IS NOT NULL AND tools_breakdown != ''
  AND timestamp >= datetime('now', ? || ' days');

-- name: GetTopProject :one
SELECT COALESCE(working_directory, 'Unknown') as directory, COUNT(*) as sessions
FROM sessions
WHERE working_directory IS NOT NULL AND working_directory != ''
  AND timestamp >= datetime('now', '-7 days')
GROUP BY working_directory
ORDER BY sessions DESC
LIMIT 1;

-- name: GetProjectMetrics :many
SELECT
    COALESCE(working_directory, 'Unknown') as directory,
    COUNT(*) as sessions,
    COALESCE(SUM(estimated_cost_usd), 0) as cost,
    COALESCE(SUM(tool_calls), 0) as tool_calls,
    COALESCE(SUM(user_prompts), 0) as prompts,
    COALESCE(AVG(duration_seconds), 0) as avg_duration
FROM sessions
WHERE working_directory IS NOT NULL AND working_directory != ''
  AND timestamp >= datetime('now', ? || ' days')
GROUP BY working_directory
ORDER BY cost DESC;

-- name: GetCacheMetrics :one
SELECT
    COALESCE(SUM(cache_read_tokens), 0) as cache_read,
    COALESCE(SUM(cache_write_tokens), 0) as cache_write,
    COALESCE(SUM(input_tokens + output_tokens), 0) as total_tokens
FROM sessions
WHERE timestamp >= datetime('now', ? || ' days');

-- name: GetCacheMetricsDaily :many
SELECT
    date(timestamp) as period,
    COALESCE(SUM(cache_read_tokens), 0) as cache_read,
    COALESCE(SUM(input_tokens + output_tokens), 0) as total_tokens
FROM sessions
WHERE timestamp >= datetime('now', ? || ' days')
GROUP BY date(timestamp)
ORDER BY period ASC;

-- name: GetEfficiencyMetrics :one
SELECT
    COALESCE(AVG(user_prompts), 0) as avg_prompts_per_session,
    COALESCE(AVG(tool_calls), 0) as avg_tools_per_session,
    COALESCE(AVG(CAST(errors_count AS REAL) / NULLIF(tool_calls, 0)), 0) as error_rate,
    COALESCE(AVG(duration_seconds), 0) as avg_duration
FROM sessions
WHERE timestamp >= datetime('now', ? || ' days');

-- name: GetEfficiencyMetricsDaily :many
SELECT
    date(timestamp) as period,
    COALESCE(AVG(user_prompts), 0) as avg_prompts,
    COALESCE(AVG(tool_calls), 0) as avg_tools,
    COUNT(*) as sessions
FROM sessions
WHERE timestamp >= datetime('now', ? || ' days')
GROUP BY date(timestamp)
ORDER BY period ASC;

-- name: GetDayOfWeekDistribution :many
SELECT
    CAST(strftime('%w', timestamp) AS INTEGER) as day_of_week,
    COUNT(*) as sessions,
    COALESCE(SUM(estimated_cost_usd), 0) as cost,
    COALESCE(AVG(user_prompts), 0) as avg_prompts
FROM sessions
WHERE timestamp >= datetime('now', ? || ' days')
GROUP BY strftime('%w', timestamp)
ORDER BY day_of_week ASC;

-- name: GetModelEfficiency :many
SELECT
    COALESCE(model, 'Unknown') as model,
    COUNT(*) as sessions,
    COALESCE(SUM(estimated_cost_usd), 0) as total_cost,
    COALESCE(SUM(input_tokens + output_tokens), 0) as total_tokens,
    CAST(
        CASE
            WHEN SUM(input_tokens + output_tokens) > 0
            THEN COALESCE(SUM(estimated_cost_usd), 0) * 1000000.0 / SUM(input_tokens + output_tokens)
            ELSE 0
        END
    AS INTEGER) as cost_per_million_tokens
FROM sessions
WHERE timestamp >= datetime('now', ? || ' days')
GROUP BY model
ORDER BY total_cost DESC;

-- ============================================================================
-- TAG & QUALITY DATA QUERIES
-- ============================================================================

-- name: AddSessionTag :exec
INSERT INTO session_tags (session_id, tag_name) VALUES (?, ?);

-- name: GetAllTags :many
SELECT name, category, color FROM tags ORDER BY category, name;

-- name: GetSessionTags :many
SELECT t.name, t.category, t.color
FROM tags t
JOIN session_tags st ON t.name = st.tag_name
WHERE st.session_id = ?
ORDER BY t.category, t.name;

