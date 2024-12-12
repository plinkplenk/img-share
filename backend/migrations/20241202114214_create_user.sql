-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users(
    id UUID UNIQUE NOT NULL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false
);
CREATE INDEX IF NOT EXISTS users_id_and_email_index ON users(id, email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS users_id_and_email_index;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
