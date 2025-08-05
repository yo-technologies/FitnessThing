-- +goose Up
ALTER TABLE llm_settings
ADD COLUMN hash VARCHAR(64) NULL;
UPDATE llm_settings
SET hash = '';
-- +goose Down
ALTER TABLE llm_settings DROP COLUMN hash;