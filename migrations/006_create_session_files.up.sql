CREATE TABLE session_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    operation TEXT NOT NULL,
    operation_count INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_session_files_session_id ON session_files(session_id);
CREATE UNIQUE INDEX idx_session_files_unique ON session_files(session_id, file_path, operation);
