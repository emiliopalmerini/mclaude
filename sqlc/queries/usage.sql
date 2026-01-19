-- name: GetRollingWindowUsage :one
SELECT
    CAST(COALESCE(SUM(m.token_cache_read + m.token_cache_write), 0) AS REAL) as total_tokens,
    CAST(COALESCE(SUM(m.cost_estimate_usd), 0) AS REAL) as total_cost
FROM sessions s
JOIN session_metrics m ON s.id = m.session_id
WHERE datetime(s.started_at) >= datetime((
    SELECT COALESCE(window_start_time, datetime('now', ? || ' hours'))
    FROM plan_config WHERE id = 1
));

-- name: UpdateWindowStartTime :exec
UPDATE plan_config SET window_start_time = ?, updated_at = datetime('now') WHERE id = 1;

-- name: UpsertPlanConfig :exec
INSERT INTO plan_config (id, plan_type, window_hours, learned_token_limit, learned_at, updated_at)
VALUES (1, ?, ?, ?, ?, datetime('now'))
ON CONFLICT (id) DO UPDATE SET
    plan_type = excluded.plan_type,
    window_hours = excluded.window_hours,
    learned_token_limit = excluded.learned_token_limit,
    learned_at = excluded.learned_at,
    updated_at = datetime('now');

-- name: GetPlanConfig :one
SELECT * FROM plan_config WHERE id = 1;

-- name: UpdateLearnedLimit :exec
UPDATE plan_config
SET learned_token_limit = ?, learned_at = datetime('now'), updated_at = datetime('now')
WHERE id = 1;

-- name: GetWeeklyWindowUsage :one
SELECT
    CAST(COALESCE(SUM(m.token_cache_read + m.token_cache_write), 0) AS REAL) as total_tokens,
    CAST(COALESCE(SUM(m.cost_estimate_usd), 0) AS REAL) as total_cost
FROM sessions s
JOIN session_metrics m ON s.id = m.session_id
WHERE datetime(s.started_at) >= datetime((
    SELECT COALESCE(weekly_window_start_time, datetime('now', '-168 hours'))
    FROM plan_config WHERE id = 1
));

-- name: UpdateWeeklyWindowStartTime :exec
UPDATE plan_config SET weekly_window_start_time = ?, updated_at = datetime('now') WHERE id = 1;

-- name: UpdateWeeklyLearnedLimit :exec
UPDATE plan_config
SET weekly_learned_token_limit = ?, weekly_learned_at = datetime('now'), updated_at = datetime('now')
WHERE id = 1;
