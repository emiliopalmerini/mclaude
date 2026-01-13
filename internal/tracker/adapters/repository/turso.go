package repository

import (
	"claude-watcher/internal/tracker/domain"
	"database/sql"
	"encoding/json"
	"fmt"
)

// TursoRepository implements SessionRepository for Turso database
type TursoRepository struct {
	db *sql.DB
}

// NewTursoRepository creates a new Turso repository
func NewTursoRepository(db *sql.DB) *TursoRepository {
	return &TursoRepository{db: db}
}

// Save persists a session to the database
func (r *TursoRepository) Save(session domain.Session) error {
	toolsBreakdownJSON, err := json.Marshal(session.Statistics.ToolsBreakdown)
	if err != nil {
		return fmt.Errorf("marshal tools breakdown: %w", err)
	}

	filesAccessedJSON, err := json.Marshal(session.Statistics.FilesAccessed)
	if err != nil {
		return fmt.Errorf("marshal files accessed: %w", err)
	}

	filesModifiedJSON, err := json.Marshal(session.Statistics.FilesModified)
	if err != nil {
		return fmt.Errorf("marshal files modified: %w", err)
	}

	cost := domain.CalculateCost(session.Statistics)
	duration := session.Statistics.Duration()

	query := `
		INSERT INTO sessions (
			session_id, instance_id, hostname, timestamp, exit_reason,
			permission_mode, working_directory, git_branch, claude_version,
			duration_seconds, user_prompts, assistant_responses,
			tool_calls, tools_breakdown, files_accessed, files_modified,
			input_tokens, output_tokens, thinking_tokens,
			cache_read_tokens, cache_write_tokens, estimated_cost_usd,
			errors_count, model, summary, rating, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(query,
		session.SessionID,
		session.InstanceID,
		session.Hostname,
		session.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		session.ExitReason,
		session.PermissionMode,
		session.WorkingDir,
		session.Statistics.GitBranch,
		session.Statistics.ClaudeVersion,
		duration,
		session.Statistics.UserPrompts,
		session.Statistics.AssistantResponses,
		session.Statistics.ToolCalls,
		string(toolsBreakdownJSON),
		string(filesAccessedJSON),
		string(filesModifiedJSON),
		session.Statistics.InputTokens,
		session.Statistics.OutputTokens,
		session.Statistics.ThinkingTokens,
		session.Statistics.CacheReadTokens,
		session.Statistics.CacheWriteTokens,
		cost,
		session.Statistics.ErrorsCount,
		session.Statistics.Model,
		session.Statistics.Summary,
		session.Rating,
		session.Notes,
	)

	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	return nil
}

// SaveTags saves tags for a session
func (r *TursoRepository) SaveTags(sessionID string, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	query := `INSERT INTO session_tags (session_id, tag_name) VALUES (?, ?)`
	for _, tag := range tags {
		if _, err := r.db.Exec(query, sessionID, tag); err != nil {
			return fmt.Errorf("insert session tag %q: %w", tag, err)
		}
	}
	return nil
}

// GetAllTags returns all available tags
func (r *TursoRepository) GetAllTags() ([]domain.Tag, error) {
	query := `SELECT name, category, color FROM tags ORDER BY category, name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query tags: %w", err)
	}
	defer rows.Close()

	var tags []domain.Tag
	for rows.Next() {
		var t domain.Tag
		if err := rows.Scan(&t.Name, &t.Category, &t.Color); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}
	return tags, nil
}
