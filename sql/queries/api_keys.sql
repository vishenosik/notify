-- name: CreateAPIKey :one
INSERT INTO api_keys (id, client_id, name, key_hash, prefix, last4, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetAPIKeyByHash :one
SELECT *
FROM api_keys
WHERE key_hash = $1 AND archived_at IS NULL;

-- name: GetAPIKeyByClientId :one
SELECT *
FROM api_keys
WHERE client_id = $1 AND archived_at IS NULL;

-- name: ListAPIKeysByClientId :many
SELECT *
FROM api_keys
WHERE client_id = $1 AND archived_at IS NULL
ORDER BY created_at DESC;

-- name: DeleteAPIKey :execrows
UPDATE api_keys
SET archived_at = $2, updated_at = $2
WHERE id = $1 AND archived_at IS NULL;

-- name: DeleteAPIKeysByClientId :execrows
UPDATE api_keys
SET archived_at = $2, updated_at = $2
WHERE client_id = $1 AND archived_at IS NULL;
