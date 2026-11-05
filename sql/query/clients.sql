-- name: CreateClient :one
INSERT INTO clients (id, name, description, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetClientById :one
SELECT *
FROM clients
WHERE id = $1 AND archived_at IS NULL;

-- name: GetClientByName :one
SELECT *
FROM clients
WHERE name = $1 AND archived_at IS NULL;

-- name: ListClients :many
SELECT *
FROM clients
WHERE archived_at IS NULL
ORDER BY created_at DESC;

-- name: DeleteClient :execrows
UPDATE clients
SET archived_at = $2, updated_at = $2
WHERE id = $1 AND archived_at IS NULL;
