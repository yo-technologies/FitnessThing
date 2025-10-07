-- +goose Up
-- +goose StatementBegin
-- Удаляем старый foreign key constraint без каскадного удаления
ALTER TABLE llm_chats
    DROP CONSTRAINT IF EXISTS llm_chats_workout_id_fkey;

-- Добавляем новый с каскадным удалением
ALTER TABLE llm_chats
    ADD CONSTRAINT llm_chats_workout_id_fkey
    FOREIGN KEY (workout_id)
    REFERENCES workouts (id)
    ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Возвращаем constraint без каскадного удаления
ALTER TABLE llm_chats
    DROP CONSTRAINT IF EXISTS llm_chats_workout_id_fkey;

ALTER TABLE llm_chats
    ADD CONSTRAINT llm_chats_workout_id_fkey
    FOREIGN KEY (workout_id)
    REFERENCES workouts (id);
-- +goose StatementEnd
