-- Test data for Claude Watcher dashboard
-- Run with: turso db shell <db-name> < scripts/seed_test_data.sql

-- Session 1: Recent feature work
INSERT INTO sessions (
    session_id, instance_id, hostname, timestamp, exit_reason, permission_mode,
    working_directory, git_branch, claude_version, duration_seconds,
    user_prompts, assistant_responses, tool_calls, input_tokens, output_tokens,
    thinking_tokens, cache_read_tokens, cache_write_tokens, estimated_cost_usd,
    errors_count, model, summary, rating, prompt_specificity, task_completion, code_confidence
) VALUES (
    'sess_001_' || hex(randomblob(4)), 'inst_001', 'macbook-pro',
    datetime('now', '-1 hours'), 'user_exit', 'default',
    '/home/user/projects/claude-watcher', 'main', '1.0.0', 1842,
    15, 15, 87, 125000, 45000, 28000, 89000, 12000, 0.0234,
    2, 'claude-sonnet-4-20250514', 'Implemented TUI dashboard with Overview and Sessions screens',
    5, 4, 5, 5
);

-- Session 2: Bug fix yesterday
INSERT INTO sessions (
    session_id, instance_id, hostname, timestamp, exit_reason, permission_mode,
    working_directory, git_branch, claude_version, duration_seconds,
    user_prompts, assistant_responses, tool_calls, input_tokens, output_tokens,
    thinking_tokens, cache_read_tokens, cache_write_tokens, estimated_cost_usd,
    errors_count, model, summary, rating, prompt_specificity, task_completion, code_confidence
) VALUES (
    'sess_002_' || hex(randomblob(4)), 'inst_002', 'macbook-pro',
    datetime('now', '-1 days', '-3 hours'), 'user_exit', 'default',
    '/home/user/projects/api-gateway', 'fix/auth-timeout', '1.0.0', 923,
    8, 8, 34, 67000, 23000, 15000, 45000, 8000, 0.0112,
    0, 'claude-sonnet-4-20250514', 'Fixed authentication timeout issue in middleware',
    4, 5, 4, 4
);

-- Session 3: Exploration session (no rating)
INSERT INTO sessions (
    session_id, instance_id, hostname, timestamp, exit_reason, permission_mode,
    working_directory, git_branch, claude_version, duration_seconds,
    user_prompts, assistant_responses, tool_calls, input_tokens, output_tokens,
    thinking_tokens, cache_read_tokens, cache_write_tokens, estimated_cost_usd,
    errors_count, model, summary
) VALUES (
    'sess_003_' || hex(randomblob(4)), 'inst_003', 'macbook-pro',
    datetime('now', '-2 days'), 'user_exit', 'default',
    '/home/user/projects/ml-pipeline', 'main', '1.0.0', 2156,
    22, 22, 156, 234000, 78000, 45000, 167000, 23000, 0.0456,
    5, 'claude-opus-4-5-20251101', 'Explored codebase structure and data flow patterns'
);

-- Session 4: Refactoring with Opus
INSERT INTO sessions (
    session_id, instance_id, hostname, timestamp, exit_reason, permission_mode,
    working_directory, git_branch, claude_version, duration_seconds,
    user_prompts, assistant_responses, tool_calls, input_tokens, output_tokens,
    thinking_tokens, cache_read_tokens, cache_write_tokens, estimated_cost_usd,
    errors_count, model, summary, rating, prompt_specificity, task_completion, code_confidence
) VALUES (
    'sess_004_' || hex(randomblob(4)), 'inst_004', 'macbook-pro',
    datetime('now', '-2 days', '-5 hours'), 'user_exit', 'default',
    '/home/user/projects/claude-watcher', 'refactor/hexagonal', '1.0.0', 3421,
    28, 28, 245, 456000, 134000, 89000, 312000, 45000, 0.1234,
    3, 'claude-opus-4-5-20251101', 'Refactored to hexagonal architecture with ports and adapters',
    5, 5, 5, 5
);

-- Session 5: Quick docs update
INSERT INTO sessions (
    session_id, instance_id, hostname, timestamp, exit_reason, permission_mode,
    working_directory, git_branch, claude_version, duration_seconds,
    user_prompts, assistant_responses, tool_calls, input_tokens, output_tokens,
    thinking_tokens, cache_read_tokens, cache_write_tokens, estimated_cost_usd,
    errors_count, model, summary
) VALUES (
    'sess_005_' || hex(randomblob(4)), 'inst_005', 'macbook-pro',
    datetime('now', '-3 days'), 'user_exit', 'default',
    '/home/user/projects/api-gateway', 'docs/readme', '1.0.0', 312,
    3, 3, 8, 12000, 5600, 3200, 8900, 1200, 0.0023,
    0, 'claude-sonnet-4-20250514', 'Updated README with installation instructions'
);

-- Session 6: Test writing
INSERT INTO sessions (
    session_id, instance_id, hostname, timestamp, exit_reason, permission_mode,
    working_directory, git_branch, claude_version, duration_seconds,
    user_prompts, assistant_responses, tool_calls, input_tokens, output_tokens,
    thinking_tokens, cache_read_tokens, cache_write_tokens, estimated_cost_usd,
    errors_count, model, summary, rating, prompt_specificity, task_completion, code_confidence
) VALUES (
    'sess_006_' || hex(randomblob(4)), 'inst_006', 'macbook-pro',
    datetime('now', '-4 days'), 'user_exit', 'default',
    '/home/user/projects/claude-watcher', 'test/transcript', '1.0.0', 1567,
    12, 12, 67, 89000, 34000, 21000, 56000, 9000, 0.0178,
    1, 'claude-sonnet-4-20250514', 'Added comprehensive tests for transcript parser',
    4, 4, 5, 4
);

-- Session 7: Config changes
INSERT INTO sessions (
    session_id, instance_id, hostname, timestamp, exit_reason, permission_mode,
    working_directory, git_branch, claude_version, duration_seconds,
    user_prompts, assistant_responses, tool_calls, input_tokens, output_tokens,
    thinking_tokens, cache_read_tokens, cache_write_tokens, estimated_cost_usd,
    errors_count, model, summary
) VALUES (
    'sess_007_' || hex(randomblob(4)), 'inst_007', 'macbook-pro',
    datetime('now', '-5 days'), 'user_exit', 'default',
    '/home/user/projects/infra', 'main', '1.0.0', 845,
    6, 6, 23, 34000, 12000, 7800, 21000, 4500, 0.0067,
    0, 'claude-haiku-3-5-20241022', 'Updated Kubernetes deployment configs'
);

-- Session 8: Large feature with errors
INSERT INTO sessions (
    session_id, instance_id, hostname, timestamp, exit_reason, permission_mode,
    working_directory, git_branch, claude_version, duration_seconds,
    user_prompts, assistant_responses, tool_calls, input_tokens, output_tokens,
    thinking_tokens, cache_read_tokens, cache_write_tokens, estimated_cost_usd,
    errors_count, model, summary, rating, prompt_specificity, task_completion, code_confidence, notes
) VALUES (
    'sess_008_' || hex(randomblob(4)), 'inst_008', 'macbook-pro',
    datetime('now', '-5 days', '-8 hours'), 'user_exit', 'default',
    '/home/user/projects/ml-pipeline', 'feat/data-loader', '1.0.0', 4523,
    35, 35, 312, 567000, 189000, 123000, 389000, 67000, 0.1567,
    8, 'claude-opus-4-5-20251101', 'Implemented async data loader with batching support',
    3, 3, 4, 3, 'Had to iterate several times on the batching logic'
);

-- Session 9: Today's work
INSERT INTO sessions (
    session_id, instance_id, hostname, timestamp, exit_reason, permission_mode,
    working_directory, git_branch, claude_version, duration_seconds,
    user_prompts, assistant_responses, tool_calls, input_tokens, output_tokens,
    thinking_tokens, cache_read_tokens, cache_write_tokens, estimated_cost_usd,
    errors_count, model, summary
) VALUES (
    'sess_009_' || hex(randomblob(4)), 'inst_009', 'macbook-pro',
    datetime('now', '-30 minutes'), 'user_exit', 'default',
    '/home/user/projects/claude-watcher', 'main', '1.0.0', 678,
    5, 5, 28, 45000, 18000, 11000, 32000, 5600, 0.0089,
    0, 'claude-sonnet-4-20250514', 'Added session detail screen to TUI dashboard'
);

-- Session 10: Another today session
INSERT INTO sessions (
    session_id, instance_id, hostname, timestamp, exit_reason, permission_mode,
    working_directory, git_branch, claude_version, duration_seconds,
    user_prompts, assistant_responses, tool_calls, input_tokens, output_tokens,
    thinking_tokens, cache_read_tokens, cache_write_tokens, estimated_cost_usd,
    errors_count, model, summary, rating, prompt_specificity, task_completion, code_confidence
) VALUES (
    'sess_010_' || hex(randomblob(4)), 'inst_010', 'macbook-pro',
    datetime('now', '-2 hours'), 'user_exit', 'default',
    '/home/user/projects/api-gateway', 'feat/rate-limit', '1.0.0', 1234,
    10, 10, 56, 78000, 28000, 17000, 52000, 8900, 0.0145,
    1, 'claude-sonnet-4-20250514', 'Added rate limiting middleware with Redis backend',
    4, 4, 4, 5
);

-- Verify data
SELECT COUNT(*) as total_sessions FROM sessions;
SELECT SUM(estimated_cost_usd) as total_cost FROM sessions;
