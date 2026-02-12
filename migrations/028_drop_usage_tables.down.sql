CREATE TABLE IF NOT EXISTS usage_metrics (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  metric_name TEXT NOT NULL,
  value REAL NOT NULL,
  attributes TEXT,
  recorded_at TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_usage_metrics_name_recorded ON usage_metrics(metric_name, recorded_at);

CREATE TABLE IF NOT EXISTS usage_limits (
  id TEXT PRIMARY KEY,
  limit_value REAL NOT NULL,
  warn_threshold REAL DEFAULT 0.8,
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS plan_config (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  plan_type TEXT NOT NULL,
  window_hours INTEGER NOT NULL DEFAULT 5,
  learned_token_limit REAL,
  learned_at TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  window_start_time TEXT,
  weekly_window_start_time TEXT,
  weekly_learned_token_limit REAL,
  weekly_learned_at TEXT
);
