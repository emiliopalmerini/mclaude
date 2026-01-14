-- name: InsertLimitEvent :exec
INSERT INTO limit_events (
    session_id,
    event_type,
    limit_type,
    timestamp,
    message,
    tokens_used,
    cost_used
) VALUES (?, ?, ?, ?, ?, ?, ?);

-- name: GetLimitEvent :one
SELECT * FROM limit_events WHERE id = ?;

-- name: ListLimitEvents :many
SELECT * FROM limit_events
ORDER BY timestamp DESC
LIMIT ? OFFSET ?;

-- name: ListLimitEventsInRange :many
SELECT * FROM limit_events
WHERE timestamp >= ? AND timestamp <= ?
ORDER BY timestamp DESC;

-- name: CountLimitEvents :one
SELECT COUNT(*) FROM limit_events;

-- name: GetLatestLimitHit :one
SELECT * FROM limit_events
WHERE event_type = 'hit'
ORDER BY timestamp DESC
LIMIT 1;

-- name: GetLatestLimitReset :one
SELECT * FROM limit_events
WHERE event_type = 'reset'
ORDER BY timestamp DESC
LIMIT 1;
