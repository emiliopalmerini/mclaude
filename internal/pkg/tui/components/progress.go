package components

import (
	"strings"

	"claude-watcher/internal/pkg/tui/theme"
)

// Progress shows step progress using dots
type Progress struct {
	Total   int
	Current int
	styles  *theme.Styles
}

// NewProgress creates a new progress indicator
func NewProgress(total, current int) Progress {
	return Progress{
		Total:   total,
		Current: current,
		styles:  theme.Default(),
	}
}

// SetCurrent updates the current step
func (p *Progress) SetCurrent(current int) {
	p.Current = current
}

// View renders the progress indicator
func (p Progress) View() string {
	var b strings.Builder

	for i := 0; i < p.Total; i++ {
		if i < p.Current {
			b.WriteString(p.styles.ProgressActive.Render("*"))
		} else if i == p.Current {
			b.WriteString(p.styles.ProgressActive.Render("o"))
		} else {
			b.WriteString(p.styles.ProgressInactive.Render("-"))
		}
		if i < p.Total-1 {
			b.WriteString(" ")
		}
	}

	return b.String()
}
