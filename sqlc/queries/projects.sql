-- name: CreateProject :exec
INSERT INTO projects (id, path, name, created_at)
VALUES (?, ?, ?, ?);

-- name: GetProjectByID :one
SELECT * FROM projects WHERE id = ?;

-- name: GetProjectByPath :one
SELECT * FROM projects WHERE path = ?;

-- name: ListProjects :many
SELECT * FROM projects ORDER BY created_at DESC;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = ?;
