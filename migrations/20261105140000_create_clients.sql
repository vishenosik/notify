-- +goose Up
CREATE TABLE clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    updated_at BIGINT NOT NULL DEFAULT (EXTRACT(EPOCH FROM NOW()))::BIGINT,
    archived_at BIGINT
);

CREATE UNIQUE INDEX clients_name_key ON clients (name) WHERE archived_at IS NULL;
CREATE INDEX clients_archived_at_idx ON clients (archived_at);

-- +goose Down
DROP TABLE IF EXISTS clients;
