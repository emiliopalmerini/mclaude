package prompter

import (
	"bufio"
	"claude-watcher/internal/tracker/domain"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// TTYPrompter collects quality feedback from the user via TTY
type TTYPrompter struct {
	logger domain.Logger
}

// NewTTYPrompter creates a new TTY prompter
func NewTTYPrompter(logger domain.Logger) *TTYPrompter {
	return &TTYPrompter{logger: logger}
}

// CollectQualityData prompts the user for session feedback via TTY.
// Returns empty QualityData if TTY is unavailable or user skips all prompts.
func (p *TTYPrompter) CollectQualityData(tags []domain.Tag) (domain.QualityData, error) {
	// Attempt to open TTY for interactive input
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		p.logger.Debug("TTY not available, skipping quality prompts")
		return domain.QualityData{}, nil
	}
	defer tty.Close()

	reader := bufio.NewReader(tty)
	data := domain.QualityData{}

	// Print header
	fmt.Fprintln(tty)
	fmt.Fprintln(tty, "Session Quality Feedback (press Enter to skip any question)")
	fmt.Fprintln(tty)

	// Group tags by category
	tagsByCategory := groupTagsByCategory(tags)

	// Ask about each category
	categories := []string{"task_type", "architecture", "prompt_style", "outcome"}
	categoryLabels := map[string]string{
		"task_type":    "Task type",
		"architecture": "Architecture",
		"prompt_style": "Prompt style",
		"outcome":      "Outcome",
	}

	for _, category := range categories {
		categoryTags := tagsByCategory[category]
		if len(categoryTags) == 0 {
			continue
		}

		selected := p.promptTagCategory(tty, reader, categoryLabels[category], categoryTags)
		data.Tags = append(data.Tags, selected...)
	}

	// Ask for rating
	data.Rating = p.promptRating(tty, reader)

	// Ask for notes
	data.Notes = p.promptNotes(tty, reader)

	// Print confirmation if any data was collected
	if len(data.Tags) > 0 || data.Rating != nil || data.Notes != "" {
		fmt.Fprintln(tty)
		fmt.Fprintln(tty, "Saved session feedback")
	}

	return data, nil
}

func (p *TTYPrompter) promptTagCategory(tty *os.File, reader *bufio.Reader, label string, tags []domain.Tag) []string {
	// Display options
	fmt.Fprintf(tty, "%s: ", label)
	for i, tag := range tags {
		fmt.Fprintf(tty, "[%d] %s ", i+1, tag.Name)
	}
	fmt.Fprintln(tty)
	fmt.Fprint(tty, "Enter numbers (comma-separated) or press Enter to skip: ")

	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	var selected []string
	parts := strings.Split(line, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		num, err := strconv.Atoi(part)
		if err != nil || num < 1 || num > len(tags) {
			continue
		}
		selected = append(selected, tags[num-1].Name)
	}
	return selected
}

func (p *TTYPrompter) promptRating(tty *os.File, reader *bufio.Reader) *int {
	fmt.Fprintln(tty)
	fmt.Fprint(tty, "Rating (1-5, or Enter to skip): ")

	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	rating, err := strconv.Atoi(line)
	if err != nil || rating < 1 || rating > 5 {
		return nil
	}
	return &rating
}

func (p *TTYPrompter) promptNotes(tty *os.File, reader *bufio.Reader) string {
	fmt.Fprintln(tty)
	fmt.Fprint(tty, "Notes (or Enter to skip): ")

	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func groupTagsByCategory(tags []domain.Tag) map[string][]domain.Tag {
	result := make(map[string][]domain.Tag)
	for _, tag := range tags {
		result[tag.Category] = append(result[tag.Category], tag)
	}
	return result
}
