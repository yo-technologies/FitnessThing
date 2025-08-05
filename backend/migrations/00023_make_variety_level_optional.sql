-- +goose Up
ALTER TABLE llm_settings
ALTER COLUMN variety_level DROP NOT NULL;
-- +goose Down
ALTER TABLE llm_settings
ALTER COLUMN variety_level
SET NOT NULL;