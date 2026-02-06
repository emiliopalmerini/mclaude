-- Add long context pricing columns for models with 1M context window
-- When input tokens exceed the threshold, premium pricing applies

ALTER TABLE model_pricing ADD COLUMN long_context_input_per_million REAL;
ALTER TABLE model_pricing ADD COLUMN long_context_output_per_million REAL;
ALTER TABLE model_pricing ADD COLUMN long_context_threshold INTEGER;

-- Add Opus 4.6 with long context pricing
INSERT INTO model_pricing (id, display_name, input_per_million, output_per_million, cache_read_per_million, cache_write_per_million, long_context_input_per_million, long_context_output_per_million, long_context_threshold, is_default)
VALUES ('claude-opus-4-6-20260206', 'Claude Opus 4.6', 5.00, 25.00, 0.50, 6.25, 10.00, 37.50, 200000, 0);

-- Update existing models with long context pricing where applicable
-- Opus 4.6 already inserted above

-- Sonnet 4.5 and Sonnet 4 get long context pricing
UPDATE model_pricing
SET long_context_input_per_million = 6.00,
    long_context_output_per_million = 22.50,
    long_context_threshold = 200000
WHERE id IN ('claude-sonnet-4-5-20241022', 'claude-sonnet-4-20250514');
