-- Create the sessions table for tracking Claude Code session usage
CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL,
    instance_id TEXT NOT NULL,
    hostname TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    exit_reason TEXT,
    permission_mode TEXT,
    working_directory TEXT,
    git_branch TEXT,
    claude_version TEXT,
    duration_seconds INTEGER,
    user_prompts INTEGER DEFAULT 0,
    assistant_responses INTEGER DEFAULT 0,
    tool_calls INTEGER DEFAULT 0,
    tools_breakdown TEXT,
    files_accessed TEXT,
    files_modified TEXT,
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    thinking_tokens INTEGER DEFAULT 0,
    cache_read_tokens INTEGER DEFAULT 0,
    cache_write_tokens INTEGER DEFAULT 0,
    estimated_cost_usd REAL DEFAULT 0.0,
    errors_count INTEGER DEFAULT 0,
    model TEXT,
    summary TEXT
);

-- Create indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_sessions_timestamp ON sessions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_sessions_hostname ON sessions(hostname);
CREATE INDEX IF NOT EXISTS idx_sessions_model ON sessions(model);
