-- +goose Up
CREATE TABLE IF NOT EXISTS prompts (
    id UUID PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL,
    prompt_text TEXT NOT NULL,
    settings_hash TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users (id)
);

-- +goose Down
DROP TABLE IF EXISTS prompts;