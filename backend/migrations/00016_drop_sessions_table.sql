-- +goose NO TRANSACTION
-- +goose Up
DROP INDEX CONCURRENTLY IF EXISTS idx_session_token;

DROP TABLE IF EXISTS sessions;

-- +goose Down
CREATE TABLE sessions
(
    id         UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id    UUID        NOT NULL,
    token      TEXT        NOT NULL,
    expired_at TIMESTAMPTZ      DEFAULT NULL,
    FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE INDEX CONCURRENTLY idx_session_token ON sessions (token);
