DROP INDEX IF EXISTS idx_sessions_limit_message;
DROP INDEX IF EXISTS idx_limit_events_type;
DROP INDEX IF EXISTS idx_limit_events_timestamp;
DROP TABLE IF EXISTS limit_events;
ALTER TABLE sessions DROP COLUMN limit_message;
