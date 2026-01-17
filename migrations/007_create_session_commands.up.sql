CREATE TABLE session_commands (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    command TEXT NOT NULL,
    exit_code INTEGER,
    executed_at TEXT
);

CREATE INDEX idx_session_commands_session_id ON session_commands(session_id);
