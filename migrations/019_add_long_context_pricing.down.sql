-- Remove Opus 4.6
DELETE FROM model_pricing WHERE id = 'claude-opus-4-6-20260115';

-- SQLite doesn't support DROP COLUMN in older versions, so we recreate the table
CREATE TABLE model_pricing_backup AS SELECT id, display_name, input_per_million, output_per_million, cache_read_per_million, cache_write_per_million, is_default, created_at FROM model_pricing;

DROP TABLE model_pricing;

CREATE TABLE model_pricing (
    id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    input_per_million REAL NOT NULL,
    output_per_million REAL NOT NULL,
    cache_read_per_million REAL,
    cache_write_per_million REAL,
    is_default INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

INSERT INTO model_pricing SELECT * FROM model_pricing_backup;
DROP TABLE model_pricing_backup;

CREATE INDEX idx_model_pricing_is_default ON model_pricing(is_default);
