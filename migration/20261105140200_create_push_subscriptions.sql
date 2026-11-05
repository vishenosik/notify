-- +goose Up
CREATE TABLE push_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES clients (id),
    endpoint TEXT NOT NULL,
    p256dh TEXT NOT NULL,
    auth TEXT NOT NULL,
    user_agent TEXT,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT
);

CREATE UNIQUE INDEX push_subscriptions_client_endpoint_key ON push_subscriptions (client_id, endpoint);
CREATE INDEX push_subscriptions_client_id_idx ON push_subscriptions (client_id);
CREATE INDEX push_subscriptions_updated_at_idx ON push_subscriptions (updated_at);

-- +goose Down
DROP TABLE IF EXISTS push_subscriptions;
