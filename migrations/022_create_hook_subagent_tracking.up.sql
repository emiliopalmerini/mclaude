CREATE TABLE IF NOT EXISTS hook_subagent_tracking (
    agent_id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    agent_type TEXT NOT NULL,
    started_at TEXT NOT NULL
);
