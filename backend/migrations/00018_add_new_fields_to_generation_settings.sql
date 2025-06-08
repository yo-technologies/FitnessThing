-- +goose Up
ALTER TABLE llm_settings
    ADD COLUMN primary_goal TEXT,
    ADD COLUMN secondary_goals TEXT[],
    ADD COLUMN experience_level TEXT,
    ADD COLUMN days_per_week SMALLINT CHECK (days_per_week BETWEEN 1 AND 7),
    ADD COLUMN session_duration_minutes SMALLINT,
    ADD COLUMN injuries TEXT,
    ADD COLUMN priority_muscle_groups UUID[] REFERENCES muscle_groups(id),
    ADD COLUMN workout_type TEXT;

-- +goose Down
ALTER TABLE llm_settings
    DROP COLUMN workout_type,
    DROP COLUMN priority_muscle_groups,
    DROP COLUMN injuries,
    DROP COLUMN session_duration_minutes,
    DROP COLUMN days_per_week,
    DROP COLUMN experience_level,
    DROP COLUMN secondary_goals,
    DROP COLUMN primary_goal;
