-- Drop indexes first
DROP INDEX IF EXISTS idx_sessions_model;
DROP INDEX IF EXISTS idx_sessions_hostname;
DROP INDEX IF EXISTS idx_sessions_session_id;
DROP INDEX IF EXISTS idx_sessions_timestamp;

-- Drop the sessions table
DROP TABLE IF EXISTS sessions;
