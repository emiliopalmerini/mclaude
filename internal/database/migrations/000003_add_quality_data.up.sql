-- Add quality feedback columns to sessions
ALTER TABLE sessions ADD COLUMN rating INTEGER;
ALTER TABLE sessions ADD COLUMN notes TEXT;
