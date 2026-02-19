CREATE TABLE IF NOT EXISTS session_quality (
    session_id TEXT PRIMARY KEY REFERENCES sessions(id) ON DELETE CASCADE,
    overall_rating INTEGER CHECK (overall_rating BETWEEN 1 AND 5),
    is_success INTEGER CHECK (is_success IN (0, 1)),
    accuracy_rating INTEGER CHECK (accuracy_rating BETWEEN 1 AND 5),
    helpfulness_rating INTEGER CHECK (helpfulness_rating BETWEEN 1 AND 5),
    efficiency_rating INTEGER CHECK (efficiency_rating BETWEEN 1 AND 5),
    notes TEXT,
    reviewed_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_session_quality_reviewed_at ON session_quality(reviewed_at);

CREATE TABLE IF NOT EXISTS session_transcripts (
    session_id TEXT PRIMARY KEY,
    gzip_data BLOB NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
