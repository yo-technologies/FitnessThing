-- +goose Up
ALTER TABLE exercise_logs ADD COLUMN "order" INT NULL;

-- +goose Down
ALTER TABLE exercise_logs DROP COLUMN "order";
