-- name: CreateExperiment :exec
INSERT INTO experiments (id, name, description, hypothesis, started_at, ended_at, is_active, created_at, model_id, plan_type, notes)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: GetExperimentByID :one
SELECT * FROM experiments WHERE id = ?;

-- name: GetExperimentByName :one
SELECT * FROM experiments WHERE name = ?;

-- name: GetActiveExperiment :one
SELECT * FROM experiments WHERE is_active = 1 LIMIT 1;

-- name: ListExperiments :many
SELECT * FROM experiments ORDER BY created_at DESC;

-- name: UpdateExperiment :exec
UPDATE experiments
SET name = ?, description = ?, hypothesis = ?, started_at = ?, ended_at = ?, is_active = ?, model_id = ?, plan_type = ?, notes = ?
WHERE id = ?;

-- name: DeleteExperiment :exec
DELETE FROM experiments WHERE id = ?;

-- name: ActivateExperiment :exec
UPDATE experiments SET is_active = 1 WHERE id = ?;

-- name: DeactivateExperiment :exec
UPDATE experiments SET is_active = 0 WHERE id = ?;

-- name: DeactivateAllExperiments :exec
UPDATE experiments SET is_active = 0;

-- name: UpsertExperimentVariable :exec
INSERT INTO experiment_variables (experiment_id, key, value)
VALUES (?, ?, ?)
ON CONFLICT(experiment_id, key) DO UPDATE SET value = excluded.value;

-- name: ListExperimentVariablesByExperimentID :many
SELECT * FROM experiment_variables WHERE experiment_id = ? ORDER BY key;

-- name: DeleteExperimentVariable :exec
DELETE FROM experiment_variables WHERE experiment_id = ? AND key = ?;
