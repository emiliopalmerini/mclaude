-- sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL UNIQUE,
    instance_id TEXT,
    hostname TEXT,
    timestamp DATETIME NOT NULL,
    working_directory TEXT,
    git_branch TEXT,
    model TEXT,
    claude_version TEXT,
    exit_reason TEXT,
    permission_mode TEXT,

    -- Interaction metrics
    user_prompts INTEGER DEFAULT 0,
    assistant_responses INTEGER DEFAULT 0,
    tool_calls INTEGER DEFAULT 0,
    tools_breakdown TEXT, -- JSON
    errors_count INTEGER DEFAULT 0,

    -- Token usage
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    thinking_tokens INTEGER DEFAULT 0,
    cache_read_tokens INTEGER DEFAULT 0,
    cache_write_tokens INTEGER DEFAULT 0,

    -- Cost
    estimated_cost_usd REAL DEFAULT 0,

    -- Files
    files_accessed TEXT, -- JSON
    files_modified TEXT, -- JSON

    -- Quality feedback
    prompt_specificity INTEGER, -- 1-5
    task_completion INTEGER,    -- 1-5
    code_confidence INTEGER,    -- 1-5
    rating INTEGER,             -- 1-5
    task_type TEXT,             -- feature/bugfix/refactor/exploration/docs/test/config
    notes TEXT,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- limit_events table
CREATE TABLE IF NOT EXISTS limit_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT,
    event_type TEXT NOT NULL, -- 'hit' or 'reset'
    limit_type TEXT NOT NULL, -- 'daily' or 'weekly'
    timestamp DATETIME NOT NULL,
    message TEXT,

    -- Usage at time of event
    tokens_used INTEGER,
    cost_used REAL,

    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (session_id) REFERENCES sessions(session_id)
);

CREATE INDEX IF NOT EXISTS idx_sessions_timestamp ON sessions(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_sessions_session_id ON sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_limit_events_timestamp ON limit_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_limit_events_session_id ON limit_events(session_id);
