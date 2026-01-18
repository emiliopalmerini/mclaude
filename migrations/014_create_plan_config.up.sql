CREATE TABLE plan_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    plan_type TEXT NOT NULL,
    window_hours INTEGER NOT NULL DEFAULT 5,
    learned_token_limit REAL,
    learned_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
