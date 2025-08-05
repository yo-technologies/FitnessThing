-- +goose Up
-- Add has_completed_onboarding column to users table
ALTER TABLE users ADD COLUMN has_completed_onboarding BOOLEAN DEFAULT FALSE;

-- Create index for performance
CREATE INDEX idx_users_has_completed_onboarding ON users(has_completed_onboarding);

-- +goose Down
-- Remove has_completed_onboarding column from users table
DROP INDEX IF EXISTS idx_users_has_completed_onboarding;
ALTER TABLE users DROP COLUMN IF EXISTS has_completed_onboarding;
