-- +goose Up
ALTER TABLE exercise_logs DROP COLUMN "order";

-- +goose Down
ALTER TABLE exercise_logs ADD COLUMN "order" INT NULL;
