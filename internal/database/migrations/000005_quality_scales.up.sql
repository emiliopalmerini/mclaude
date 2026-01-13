-- Add new scale columns for quality tracking
ALTER TABLE sessions ADD COLUMN prompt_specificity INTEGER;
ALTER TABLE sessions ADD COLUMN task_completion INTEGER;
ALTER TABLE sessions ADD COLUMN code_confidence INTEGER;

-- Remove architecture tags (project-level, not session-level)
DELETE FROM session_tags WHERE tag_name IN (
    'vertical-slice', 'hexagonal', 'mvc', 'solid', 'ddd'
);
DELETE FROM tags WHERE category = 'architecture';

-- Remove prompt_style tags (replaced by prompt_specificity scale)
DELETE FROM session_tags WHERE tag_name IN (
    'detailed-upfront', 'iterative', 'minimal'
);
DELETE FROM tags WHERE category = 'prompt_style';

-- Remove outcome tags (replaced by task_completion scale)
DELETE FROM session_tags WHERE tag_name IN (
    'success', 'partial', 'failed', 'rework-needed'
);
DELETE FROM tags WHERE category = 'outcome';
