-- name: CreateUsageMetric :exec
INSERT INTO usage_metrics (metric_name, value, attributes, recorded_at)
VALUES (?, ?, ?, ?);

-- name: GetUsageForPeriod :many
SELECT
    metric_name,
    SUM(value) as total_value
FROM usage_metrics
WHERE recorded_at >= ? AND recorded_at < ?
GROUP BY metric_name;

-- name: GetTokenUsageByType :many
SELECT
    json_extract(attributes, '$.type') as token_type,
    SUM(value) as total
FROM usage_metrics
WHERE metric_name = 'claude_code.token.usage'
  AND recorded_at >= ?
GROUP BY token_type;

-- name: GetDailyUsageSummary :one
SELECT
    CAST(COALESCE(SUM(CASE WHEN metric_name = 'claude_code.token.usage' THEN value ELSE 0 END), 0) AS REAL) as total_tokens,
    CAST(COALESCE(SUM(CASE WHEN metric_name = 'claude_code.cost.usage' THEN value ELSE 0 END), 0) AS REAL) as total_cost
FROM usage_metrics
WHERE recorded_at >= date('now', 'start of day');

-- name: GetWeeklyUsageSummary :one
SELECT
    CAST(COALESCE(SUM(CASE WHEN metric_name = 'claude_code.token.usage' THEN value ELSE 0 END), 0) AS REAL) as total_tokens,
    CAST(COALESCE(SUM(CASE WHEN metric_name = 'claude_code.cost.usage' THEN value ELSE 0 END), 0) AS REAL) as total_cost
FROM usage_metrics
WHERE recorded_at >= date('now', 'weekday 0', '-7 days');

-- name: DeleteUsageMetricsBefore :execrows
DELETE FROM usage_metrics WHERE recorded_at < ?;

-- name: CreateUsageLimit :exec
INSERT INTO usage_limits (id, limit_value, warn_threshold, enabled, updated_at)
VALUES (?, ?, ?, ?, datetime('now'))
ON CONFLICT (id) DO UPDATE SET
    limit_value = excluded.limit_value,
    warn_threshold = excluded.warn_threshold,
    enabled = excluded.enabled,
    updated_at = datetime('now');

-- name: GetUsageLimit :one
SELECT * FROM usage_limits WHERE id = ?;

-- name: ListUsageLimits :many
SELECT * FROM usage_limits ORDER BY id;

-- name: DeleteUsageLimit :exec
DELETE FROM usage_limits WHERE id = ?;

-- name: GetRollingWindowUsage :one
SELECT
    CAST(COALESCE(SUM(CASE WHEN metric_name = 'claude_code.token.usage' THEN value ELSE 0 END), 0) AS REAL) as total_tokens,
    CAST(COALESCE(SUM(CASE WHEN metric_name = 'claude_code.cost.usage' THEN value ELSE 0 END), 0) AS REAL) as total_cost
FROM usage_metrics
WHERE recorded_at >= datetime('now', ? || ' hours');

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
