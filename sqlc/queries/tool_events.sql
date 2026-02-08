-- name: ListToolEventsBySessionID :many
SELECT * FROM tool_events WHERE session_id = ? ORDER BY captured_at ASC;
