-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS llm_chats (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    workout_id UUID,
    title TEXT,
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (workout_id) REFERENCES workouts (id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_llm_chats_workout_id
    ON llm_chats (workout_id)
    WHERE workout_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_llm_chats_user_id_created_at
    ON llm_chats (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS llm_chat_messages (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    chat_id UUID NOT NULL REFERENCES llm_chats (id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    tool_name TEXT,
    tool_call_id TEXT,
    tool_arguments TEXT,
    token_usage INTEGER,
    error TEXT
);

CREATE INDEX IF NOT EXISTS idx_llm_chat_messages_chat_id_created_at
    ON llm_chat_messages (chat_id, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS llm_chat_messages;
DROP TABLE IF EXISTS llm_chats;
-- +goose StatementEnd
