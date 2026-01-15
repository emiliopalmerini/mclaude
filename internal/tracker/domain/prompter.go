package domain

// QualityData holds user feedback collected after a session ends
type QualityData struct {
	Tags              []string // Selected tag names (task_type only)
	Rating            *int     // 1-5 session satisfaction, nil if skipped
	PromptSpecificity *int     // 1-5 how detailed prompts were
	TaskCompletion    *int     // 1-5 how complete the work is
	CodeConfidence    *int     // 1-5 confidence in generated code
	Notes             string   // Free-form notes
}

// Tag represents a session tag with its category and display color
type Tag struct {
	Name     string
	Category string
	Color    string
}

// DefaultTaskTypeTags returns the predefined task type tags
func DefaultTaskTypeTags() []Tag {
	return []Tag{
		{Name: "feature", Category: "task_type", Color: "#10B981"},
		{Name: "bugfix", Category: "task_type", Color: "#EF4444"},
		{Name: "refactor", Category: "task_type", Color: "#8B5CF6"},
		{Name: "exploration", Category: "task_type", Color: "#3B82F6"},
		{Name: "docs", Category: "task_type", Color: "#F59E0B"},
		{Name: "test", Category: "task_type", Color: "#EC4899"},
		{Name: "config", Category: "task_type", Color: "#6B7280"},
	}
}

// Prompter collects quality feedback from the user
type Prompter interface {
	// CollectQualityData prompts the user for session feedback.
	// Returns empty QualityData if TTY is unavailable or user skips all prompts.
	CollectQualityData(tags []Tag) (QualityData, error)
}
