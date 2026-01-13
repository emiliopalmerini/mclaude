-- Remove scale columns
ALTER TABLE sessions DROP COLUMN prompt_specificity;
ALTER TABLE sessions DROP COLUMN task_completion;
ALTER TABLE sessions DROP COLUMN code_confidence;

-- Restore architecture tags
INSERT INTO tags (name, category, color) VALUES ('vertical-slice', 'architecture', '#10b981');
INSERT INTO tags (name, category, color) VALUES ('hexagonal', 'architecture', '#6366f1');
INSERT INTO tags (name, category, color) VALUES ('mvc', 'architecture', '#f97316');
INSERT INTO tags (name, category, color) VALUES ('solid', 'architecture', '#ec4899');
INSERT INTO tags (name, category, color) VALUES ('ddd', 'architecture', '#14b8a6');

-- Restore prompt_style tags
INSERT INTO tags (name, category, color) VALUES ('detailed-upfront', 'prompt_style', '#a855f7');
INSERT INTO tags (name, category, color) VALUES ('iterative', 'prompt_style', '#0ea5e9');
INSERT INTO tags (name, category, color) VALUES ('minimal', 'prompt_style', '#84cc16');

-- Restore outcome tags
INSERT INTO tags (name, category, color) VALUES ('success', 'outcome', '#22c55e');
INSERT INTO tags (name, category, color) VALUES ('partial', 'outcome', '#eab308');
INSERT INTO tags (name, category, color) VALUES ('failed', 'outcome', '#ef4444');
INSERT INTO tags (name, category, color) VALUES ('rework-needed', 'outcome', '#f97316');
