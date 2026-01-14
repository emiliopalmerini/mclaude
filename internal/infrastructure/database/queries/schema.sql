-- Schema for sqlc type generation (mirrors actual database schema)
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY,
    session_id TEXT NOT NULL UNIQUE,
    instance_id TEXT,
    hostname TEXT,
    timestamp TEXT NOT NULL,
    working_directory TEXT,
    git_branch TEXT,
    model TEXT,
    claude_version TEXT,
    exit_reason TEXT,
    permission_mode TEXT,

    -- Interaction metrics
    user_prompts INTEGER,
    assistant_responses INTEGER,
    tool_calls INTEGER,
    tools_breakdown TEXT, -- JSON
    errors_count INTEGER,

    -- Token usage
    input_tokens INTEGER,
    output_tokens INTEGER,
    thinking_tokens INTEGER,
    cache_read_tokens INTEGER,
    cache_write_tokens INTEGER,

    -- Cost
    estimated_cost_usd REAL,

    -- Files
    files_accessed TEXT, -- JSON
    files_modified TEXT, -- JSON

    -- Quality feedback
    prompt_specificity INTEGER,
    task_completion INTEGER,
    code_confidence INTEGER,
    rating INTEGER,
    task_type TEXT,
    notes TEXT,

    created_at TEXT
);

CREATE TABLE limit_events (
    id INTEGER PRIMARY KEY,
    session_id TEXT,
    event_type TEXT NOT NULL,
    limit_type TEXT NOT NULL,
    timestamp TEXT NOT NULL,
    message TEXT,
    tokens_used INTEGER,
    cost_used REAL,
    created_at TEXT
);
