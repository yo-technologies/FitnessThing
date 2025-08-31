-- +goose Up
ALTER TABLE workouts
ADD COLUMN IF NOT EXISTS is_generating BOOLEAN DEFAULT FALSE;
ALTER TABLE workouts
ADD COLUMN IF NOT EXISTS generation_error TEXT;

-- +goose Down
ALTER TABLE workouts DROP COLUMN IF EXISTS is_generating;
ALTER TABLE workouts DROP COLUMN IF EXISTS generation_error;
