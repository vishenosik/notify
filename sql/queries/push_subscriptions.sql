-- name: UpsertPushSubscription :one
INSERT INTO push_subscriptions (id, client_id, endpoint, p256dh, auth, user_agent, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (client_id, endpoint) DO UPDATE
SET
    p256dh = EXCLUDED.p256dh,
    auth = EXCLUDED.auth,
    user_agent = EXCLUDED.user_agent,
    updated_at = EXCLUDED.updated_at
RETURNING *;

-- name: GetPushSubscriptionById :one
SELECT *
FROM push_subscriptions
WHERE id = $1;

-- name: GetPushSubscriptionByClientIdAndEndpoint :one
SELECT *
FROM push_subscriptions
WHERE client_id = $1 AND endpoint = $2;

-- name: ListPushSubscriptionsByClientId :many
SELECT *
FROM push_subscriptions
WHERE client_id = $1
ORDER BY created_at ASC;

-- name: DeletePushSubscriptionById :exec
DELETE FROM push_subscriptions
WHERE id = $1;

-- name: DeletePushSubscriptionByClientIdAndEndpoint :exec
DELETE FROM push_subscriptions
WHERE client_id = $1 AND endpoint = $2;

-- name: DeletePushSubscriptionsByClientId :exec
DELETE FROM push_subscriptions
WHERE client_id = $1;
