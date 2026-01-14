-- name: InsertSession :exec
INSERT INTO sessions (
    session_id,
    instance_id,
    hostname,
    timestamp,
    working_directory,
    git_branch,
    model,
    claude_version,
    exit_reason,
    permission_mode,
    user_prompts,
    assistant_responses,
    tool_calls,
    tools_breakdown,
    errors_count,
    input_tokens,
    output_tokens,
    thinking_tokens,
    cache_read_tokens,
    cache_write_tokens,
    estimated_cost_usd,
    files_accessed,
    files_modified,
    prompt_specificity,
    task_completion,
    code_confidence,
    rating,
    task_type,
    notes
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetSession :one
SELECT * FROM sessions WHERE session_id = ?;

-- name: ListSessions :many
SELECT * FROM sessions
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: ListSessionsInRange :many
SELECT * FROM sessions
WHERE timestamp >= ? AND timestamp <= ?
ORDER BY timestamp DESC;

-- name: CountSessions :one
SELECT COUNT(*) FROM sessions;

-- name: CountSessionsInRange :one
SELECT COUNT(*) FROM sessions
WHERE timestamp >= ? AND timestamp <= ?;

-- name: GetTotalTokensInRange :one
SELECT
    COALESCE(SUM(input_tokens), 0) as input_tokens,
    COALESCE(SUM(output_tokens), 0) as output_tokens,
    COALESCE(SUM(thinking_tokens), 0) as thinking_tokens,
    COALESCE(SUM(cache_read_tokens), 0) as cache_read_tokens,
    COALESCE(SUM(cache_write_tokens), 0) as cache_write_tokens
FROM sessions
WHERE timestamp >= ? AND timestamp <= ?;

-- name: GetTotalCostInRange :one
SELECT COALESCE(SUM(estimated_cost_usd), 0) as total_cost
FROM sessions
WHERE timestamp >= ? AND timestamp <= ?;

-- name: GetAverageRatingsInRange :one
SELECT
    AVG(rating) as avg_rating,
    AVG(prompt_specificity) as avg_prompt_specificity,
    AVG(task_completion) as avg_task_completion,
    AVG(code_confidence) as avg_code_confidence
FROM sessions
WHERE timestamp >= ? AND timestamp <= ?
  AND (rating IS NOT NULL OR prompt_specificity IS NOT NULL
       OR task_completion IS NOT NULL OR code_confidence IS NOT NULL);
