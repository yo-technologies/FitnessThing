-- +goose Up
ALTER TABLE workouts
    ADD COLUMN IF NOT EXISTS generation_status TEXT; -- '', 'running', 'failed', 'completed'

-- migrate existing state: if is_generating=true -> running, else if reasoning<>'' -> completed
UPDATE workouts SET generation_status = 'running' WHERE is_generating = true;
UPDATE workouts SET generation_status = 'completed' WHERE (is_generating = false OR is_generating IS NULL) AND COALESCE(reasoning,'') <> '';

-- drop old boolean column
ALTER TABLE workouts DROP COLUMN IF EXISTS is_generating;

-- generation_error остается (внутреннее)

-- +goose Down
ALTER TABLE workouts ADD COLUMN IF NOT EXISTS is_generating BOOLEAN DEFAULT FALSE;
-- восстановить булево из статуса
UPDATE workouts SET is_generating = (generation_status = 'running');
ALTER TABLE workouts DROP COLUMN IF EXISTS generation_status;