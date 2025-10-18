-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS llm_token_usage (
    user_id UUID NOT NULL REFERENCES users (id),
    day DATE NOT NULL,
    used_tokens INTEGER NOT NULL DEFAULT 0,
    reserved_tokens INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, day)
);

CREATE INDEX IF NOT EXISTS idx_llm_token_usage_user_day ON llm_token_usage (user_id, day);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS llm_token_usage;
-- +goose StatementEnd
