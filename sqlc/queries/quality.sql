-- name: UpsertSessionQuality :exec
INSERT INTO session_quality (
    session_id, overall_rating, is_success, accuracy_rating,
    helpfulness_rating, efficiency_rating, notes, reviewed_at, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
ON CONFLICT (session_id) DO UPDATE SET
    overall_rating = excluded.overall_rating,
    is_success = excluded.is_success,
    accuracy_rating = excluded.accuracy_rating,
    helpfulness_rating = excluded.helpfulness_rating,
    efficiency_rating = excluded.efficiency_rating,
    notes = excluded.notes,
    reviewed_at = excluded.reviewed_at;

-- name: GetSessionQualityBySessionID :one
SELECT * FROM session_quality WHERE session_id = ?;

-- name: DeleteSessionQuality :exec
DELETE FROM session_quality WHERE session_id = ?;

-- name: ListSessionQualitiesForSessions :many
SELECT session_id, overall_rating, is_success, reviewed_at
FROM session_quality
WHERE reviewed_at IS NOT NULL;

-- name: GetOverallQualityStats :one
SELECT
    COUNT(DISTINCT sq.session_id) as reviewed_count,
    AVG(sq.overall_rating) as avg_overall_rating,
    SUM(CASE WHEN sq.is_success = 1 THEN 1 ELSE 0 END) as success_count,
    SUM(CASE WHEN sq.is_success = 0 THEN 1 ELSE 0 END) as failure_count,
    AVG(sq.accuracy_rating) as avg_accuracy,
    AVG(sq.helpfulness_rating) as avg_helpfulness,
    AVG(sq.efficiency_rating) as avg_efficiency
FROM session_quality sq
WHERE sq.reviewed_at IS NOT NULL;

-- name: ListUnreviewedSessionIDs :many
SELECT s.id FROM sessions s
LEFT JOIN session_quality sq ON s.id = sq.session_id
WHERE sq.reviewed_at IS NULL OR sq.session_id IS NULL
ORDER BY s.created_at DESC
LIMIT ?;

-- name: GetQualityStatsByExperiment :one
SELECT
    COUNT(DISTINCT sq.session_id) as reviewed_count,
    AVG(sq.overall_rating) as avg_overall_rating,
    SUM(CASE WHEN sq.is_success = 1 THEN 1 ELSE 0 END) as success_count,
    SUM(CASE WHEN sq.is_success = 0 THEN 1 ELSE 0 END) as failure_count,
    AVG(sq.accuracy_rating) as avg_accuracy,
    AVG(sq.helpfulness_rating) as avg_helpfulness,
    AVG(sq.efficiency_rating) as avg_efficiency
FROM sessions s
JOIN session_quality sq ON s.id = sq.session_id
WHERE s.experiment_id = ? AND sq.reviewed_at IS NOT NULL;

-- name: GetQualityStatsForAllExperiments :many
SELECT
    e.id as experiment_id,
    e.name as experiment_name,
    COUNT(DISTINCT sq.session_id) as reviewed_count,
    AVG(sq.overall_rating) as avg_overall_rating,
    SUM(CASE WHEN sq.is_success = 1 THEN 1 ELSE 0 END) as success_count,
    SUM(CASE WHEN sq.is_success = 0 THEN 1 ELSE 0 END) as failure_count,
    AVG(sq.accuracy_rating) as avg_accuracy,
    AVG(sq.helpfulness_rating) as avg_helpfulness,
    AVG(sq.efficiency_rating) as avg_efficiency
FROM experiments e
LEFT JOIN sessions s ON s.experiment_id = e.id
LEFT JOIN session_quality sq ON s.id = sq.session_id AND sq.reviewed_at IS NOT NULL
GROUP BY e.id, e.name;
