-- +goose Up
ALTER TABLE workouts DROP COLUMN IF EXISTS is_ai_generated;
ALTER TABLE workouts DROP COLUMN IF EXISTS reasoning;
ALTER TABLE workouts DROP COLUMN IF EXISTS generation_status;
ALTER TABLE workouts DROP COLUMN IF EXISTS generation_error;

-- +goose Down
ALTER TABLE workouts ADD COLUMN IF NOT EXISTS is_ai_generated BOOLEAN DEFAULT FALSE;
ALTER TABLE workouts ADD COLUMN IF NOT EXISTS reasoning TEXT;
ALTER TABLE workouts ADD COLUMN IF NOT EXISTS generation_status TEXT;
ALTER TABLE workouts ADD COLUMN IF NOT EXISTS generation_error TEXT;
