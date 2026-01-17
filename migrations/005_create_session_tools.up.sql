CREATE TABLE session_tools (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    tool_name TEXT NOT NULL,
    invocation_count INTEGER NOT NULL DEFAULT 1,
    total_duration_ms INTEGER,
    error_count INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_session_tools_session_id ON session_tools(session_id);
CREATE INDEX idx_session_tools_tool_name ON session_tools(tool_name);
CREATE UNIQUE INDEX idx_session_tools_unique ON session_tools(session_id, tool_name);
