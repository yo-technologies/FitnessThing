-- +goose Up
ALTER TABLE llm_settings
ALTER COLUMN base_prompt DROP NOT NULL;

-- +goose Down
ALTER TABLE llm_settings
ALTER COLUMN base_prompt SET NOT NULL;
