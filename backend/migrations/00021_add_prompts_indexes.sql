-- +goose NO TRANSACTION
-- +goose Up
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_prompts_user_id ON prompts (user_id);

-- +goose Down
DROP INDEX CONCURRENTLY IF EXISTS idx_prompts_user_id;
