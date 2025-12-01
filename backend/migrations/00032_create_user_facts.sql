-- +goose Up
CREATE TABLE IF NOT EXISTS user_facts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_facts_user_id ON user_facts(user_id);

-- +goose Down
DROP TABLE IF EXISTS user_facts;