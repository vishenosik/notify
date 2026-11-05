-- +goose Up
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL REFERENCES clients (id),
    name VARCHAR(255) NOT NULL,
    key_hash BYTEA NOT NULL UNIQUE,
    prefix VARCHAR(8) NOT NULL,
    last4 VARCHAR(4) NOT NULL,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    archived_at BIGINT
);

CREATE INDEX api_keys_client_id_idx ON api_keys (client_id);
CREATE INDEX api_keys_archived_at_idx ON api_keys (archived_at);

-- +goose Down
DROP TABLE IF EXISTS api_keys;
