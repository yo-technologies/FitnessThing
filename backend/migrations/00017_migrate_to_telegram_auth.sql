-- +goose NO TRANSACTION
-- +goose Up
ALTER TABLE users ADD COLUMN telegram_id BIGINT;
ALTER TABLE users ADD COLUMN username VARCHAR(255);

UPDATE users SET telegram_id = -1 * ABS(EXTRACT(EPOCH FROM created_at)::BIGINT) WHERE telegram_id IS NULL;

ALTER TABLE users ALTER COLUMN telegram_id SET NOT NULL;

ALTER TABLE users ADD CONSTRAINT users_telegram_id_unique UNIQUE (telegram_id);
ALTER TABLE users ADD CONSTRAINT users_username_unique UNIQUE (username);

CREATE INDEX CONCURRENTLY idx_users_telegram_id ON users (telegram_id);
CREATE INDEX CONCURRENTLY idx_users_username ON users (username) WHERE username IS NOT NULL;

DROP INDEX IF EXISTS idx_users_email;

ALTER TABLE users DROP COLUMN email;
ALTER TABLE users DROP COLUMN password;

-- +goose Down
ALTER TABLE users ADD COLUMN email VARCHAR(255);
ALTER TABLE users ADD COLUMN password VARCHAR(255);

UPDATE users SET email = 'user_' || id || '@telegram.temp', password = 'temp_password' WHERE email IS NULL;

ALTER TABLE users ALTER COLUMN email SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);

CREATE INDEX idx_users_email ON users (email);

DROP INDEX CONCURRENTLY IF EXISTS idx_users_telegram_id;
DROP INDEX CONCURRENTLY IF EXISTS idx_users_username;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_telegram_id_unique;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_unique;

ALTER TABLE users DROP COLUMN telegram_id;
ALTER TABLE users DROP COLUMN username;
