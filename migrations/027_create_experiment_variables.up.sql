CREATE TABLE IF NOT EXISTS experiment_variables (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    experiment_id TEXT NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(experiment_id, key)
);

CREATE INDEX IF NOT EXISTS idx_experiment_variables_experiment_id ON experiment_variables(experiment_id);
