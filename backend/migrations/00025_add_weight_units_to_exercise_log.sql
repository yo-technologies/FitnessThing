-- +goose Up
ALTER TABLE exercise_logs ADD COLUMN IF NOT EXISTS weight_unit VARCHAR(8) NOT NULL DEFAULT 'kg';


-- +goose Down
ALTER TABLE exercise_logs DROP COLUMN IF EXISTS weight_unit;
