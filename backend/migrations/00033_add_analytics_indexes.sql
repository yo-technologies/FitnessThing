-- +goose Up
-- +goose StatementBegin
-- Composite index for workouts analytics query
-- Covers user_id and finished_at filtering, the most common analytics query pattern
CREATE INDEX IF NOT EXISTS idx_workouts_user_finished ON workouts (user_id, finished_at DESC)
WHERE finished_at IS NOT NULL;
-- Index for exercise_muscle_groups joins
CREATE INDEX IF NOT EXISTS idx_exercise_muscle_groups_exercise_id ON exercise_muscle_groups (exercise_id);
CREATE INDEX IF NOT EXISTS idx_exercise_muscle_groups_muscle_group_id ON exercise_muscle_groups (muscle_group_id);
-- Index for exercise_logs joins with workout_id and exercise_id
CREATE INDEX IF NOT EXISTS idx_exercise_logs_workout_exercise ON exercise_logs (workout_id, exercise_id);
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_workouts_user_finished;
DROP INDEX IF EXISTS idx_exercise_muscle_groups_exercise_id;
DROP INDEX IF EXISTS idx_exercise_muscle_groups_muscle_group_id;
DROP INDEX IF EXISTS idx_exercise_logs_workout_exercise;
-- +goose StatementEnd