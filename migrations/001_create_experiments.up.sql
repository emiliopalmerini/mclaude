CREATE TABLE experiments (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    hypothesis TEXT,
    started_at TEXT NOT NULL,
    ended_at TEXT,
    is_active INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_experiments_is_active ON experiments(is_active);
CREATE INDEX idx_experiments_name ON experiments(name);
