-- Composite index for queries joining session_tools/session_files on sessions
-- filtered by created_at (e.g. dashboard GetTopToolsUsage).
-- Covers the created_at filter and provides the session id for the join.
CREATE INDEX IF NOT EXISTS idx_sessions_created_at_id ON sessions(created_at DESC, id);
