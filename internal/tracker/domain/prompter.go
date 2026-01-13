package domain

// QualityData holds user feedback collected after a session ends
type QualityData struct {
	Tags   []string // Selected tag names
	Rating *int     // 1-5 rating, nil if skipped
	Notes  string   // Free-form notes
}

// Tag represents a session tag with its category and display color
type Tag struct {
	Name     string
	Category string
	Color    string
}

// Prompter collects quality feedback from the user
type Prompter interface {
	// CollectQualityData prompts the user for session feedback.
	// Returns empty QualityData if TTY is unavailable or user skips all prompts.
	CollectQualityData(tags []Tag) (QualityData, error)
}
