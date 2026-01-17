CREATE TABLE session_metrics (
    session_id TEXT PRIMARY KEY REFERENCES sessions(id) ON DELETE CASCADE,
    message_count_user INTEGER NOT NULL DEFAULT 0,
    message_count_assistant INTEGER NOT NULL DEFAULT 0,
    turn_count INTEGER NOT NULL DEFAULT 0,
    token_input INTEGER NOT NULL DEFAULT 0,
    token_output INTEGER NOT NULL DEFAULT 0,
    token_cache_read INTEGER NOT NULL DEFAULT 0,
    token_cache_write INTEGER NOT NULL DEFAULT 0,
    cost_estimate_usd REAL,
    error_count INTEGER NOT NULL DEFAULT 0
);
