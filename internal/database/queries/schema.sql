-- Schema for sqlc type generation only (tables already exist in Turso)
CREATE TABLE sessions (
    id INTEGER PRIMARY KEY,
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
    user_prompts INTEGER,
    assistant_responses INTEGER,
    tool_calls INTEGER,
    tools_breakdown TEXT,
    files_accessed TEXT,
    files_modified TEXT,
    input_tokens INTEGER,
    output_tokens INTEGER,
    thinking_tokens INTEGER,
    cache_read_tokens INTEGER,
    cache_write_tokens INTEGER,
    estimated_cost_usd REAL,
    errors_count INTEGER,
    model TEXT,
    summary TEXT,
    rating INTEGER,
    notes TEXT
);

CREATE TABLE tags (
    name TEXT PRIMARY KEY,
    category TEXT NOT NULL,
    color TEXT NOT NULL
);

CREATE TABLE session_tags (
    session_id TEXT NOT NULL,
    tag_name TEXT NOT NULL,
    created_at TEXT,
    PRIMARY KEY (session_id, tag_name)
);
