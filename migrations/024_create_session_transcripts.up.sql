CREATE TABLE IF NOT EXISTS session_transcripts (
    session_id TEXT PRIMARY KEY,
    gzip_data BLOB NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
