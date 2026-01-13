-- Add limit_message column to sessions to capture when a limit was hit
ALTER TABLE sessions ADD COLUMN limit_message TEXT;

-- Track limit events with usage since last limit
CREATE TABLE IF NOT EXISTS limit_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT NOT NULL,
    limit_type TEXT NOT NULL CHECK(limit_type IN ('daily', 'weekly')),
    reset_time TEXT,

    -- Usage since last limit event (deltas)
    sessions_count INTEGER DEFAULT 0,
    input_tokens INTEGER DEFAULT 0,
    output_tokens INTEGER DEFAULT 0,
    thinking_tokens INTEGER DEFAULT 0,
    total_cost_usd REAL DEFAULT 0.0
);

CREATE INDEX IF NOT EXISTS idx_limit_events_timestamp ON limit_events(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_limit_events_type ON limit_events(limit_type);
CREATE INDEX IF NOT EXISTS idx_sessions_limit_message ON sessions(limit_message) WHERE limit_message IS NOT NULL;
