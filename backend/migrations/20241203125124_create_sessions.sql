-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS auth_sessions(
    id VARCHAR(48) UNIQUE NOT NULL,
    user_id UUID,
    expires_on TIMESTAMP,
    CONSTRAINT fk_auth_sessions_user FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX IF NOT EXISTS auth_sessions_id_and_user_id_index ON auth_sessions(id, user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP INDEX IF EXISTS auth_sessions_id_and_user_id_index;
DROP TABLE IF EXISTS auth_sessions;
-- +goose StatementEnd
