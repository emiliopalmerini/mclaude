CREATE TABLE usage_limits (
    id TEXT PRIMARY KEY,
    limit_value REAL NOT NULL,
    warn_threshold REAL DEFAULT 0.8,
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
