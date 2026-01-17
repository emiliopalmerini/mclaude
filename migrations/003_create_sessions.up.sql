CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id),
    experiment_id TEXT REFERENCES experiments(id),
    transcript_path TEXT NOT NULL,
    transcript_stored_path TEXT,
    cwd TEXT NOT NULL,
    permission_mode TEXT NOT NULL,
    exit_reason TEXT NOT NULL,
    started_at TEXT,
    ended_at TEXT,
    duration_seconds INTEGER,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_sessions_project_id ON sessions(project_id);
CREATE INDEX idx_sessions_experiment_id ON sessions(experiment_id);
CREATE INDEX idx_sessions_created_at ON sessions(created_at);
CREATE INDEX idx_sessions_started_at ON sessions(started_at);
