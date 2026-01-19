-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
-- This is a destructive migration - weekly data will be lost

CREATE TABLE plan_config_new (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    plan_type TEXT NOT NULL,
    window_hours INTEGER NOT NULL DEFAULT 5,
    learned_token_limit REAL,
    learned_at TEXT,
    window_start_time TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO plan_config_new (id, plan_type, window_hours, learned_token_limit, learned_at, window_start_time, created_at, updated_at)
SELECT id, plan_type, window_hours, learned_token_limit, learned_at, window_start_time, created_at, updated_at
FROM plan_config;

DROP TABLE plan_config;
ALTER TABLE plan_config_new RENAME TO plan_config;
