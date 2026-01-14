package components

import (
	"fmt"
	"strings"

	"claude-watcher/internal/pkg/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
)

// Scale is a 1-5 scale selector component
type Scale struct {
	Label    string
	LowDesc  string
	HighDesc string
	Value    int  // 0 = not set, 1-5 = selected
	Cursor   int  // Current cursor position (1-5)
	Focused  bool // Whether this component is focused
	styles   *theme.Styles
}

// NewScale creates a new scale selector
func NewScale(label, lowDesc, highDesc string) Scale {
	return Scale{
		Label:    label,
		LowDesc:  lowDesc,
		HighDesc: highDesc,
		Value:    0,
		Cursor:   3, // Start in middle
		Focused:  false,
		styles:   theme.Default(),
	}
}

// Focus sets the scale as focused
func (s *Scale) Focus() {
	s.Focused = true
}

// Blur removes focus from the scale
func (s *Scale) Blur() {
	s.Focused = false
}

// SetValue sets the scale value directly
func (s *Scale) SetValue(v int) {
	if v >= 1 && v <= 5 {
		s.Value = v
		s.Cursor = v
	}
}

// Update handles key events for the scale
func (s Scale) Update(msg tea.Msg) (Scale, tea.Cmd) {
	if !s.Focused {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "h", "left":
			if s.Cursor > 1 {
				s.Cursor--
			}
		case "l", "right":
			if s.Cursor < 5 {
				s.Cursor++
			}
		case " ", "enter":
			s.Value = s.Cursor
		case "1", "2", "3", "4", "5":
			num := int(msg.String()[0] - '0')
			s.Cursor = num
			s.Value = num
		}
	}

	return s, nil
}

// View renders the scale
func (s Scale) View() string {
	var b strings.Builder

	b.WriteString(s.styles.Subtitle.Render(s.Label))
	b.WriteString("\n\n")

	// Scale labels
	b.WriteString("  ")
	b.WriteString(s.styles.Muted.Render(s.LowDesc))
	padding := 30 - len(s.LowDesc) - len(s.HighDesc)
	if padding < 4 {
		padding = 4
	}
	b.WriteString(strings.Repeat(" ", padding))
	b.WriteString(s.styles.Muted.Render(s.HighDesc))
	b.WriteString("\n")

	// Number row
	b.WriteString("  ")
	for i := 1; i <= 5; i++ {
		isCursor := s.Focused && i == s.Cursor
		isSelected := i == s.Value
		numStr := fmt.Sprintf("%d", i)

		if isCursor {
			if isSelected {
				b.WriteString(s.styles.Active.Render("【" + numStr + "】"))
			} else {
				b.WriteString(s.styles.Active.Render("[ " + numStr + " ]"))
			}
		} else if isSelected {
			b.WriteString(s.styles.Selected.Render("  " + numStr + "  "))
		} else {
			b.WriteString(s.styles.Muted.Render("  " + numStr + "  "))
		}
		b.WriteString(" ")
	}
	b.WriteString("\n")

	// Cursor indicator row (only when focused)
	if s.Focused {
		b.WriteString("  ")
		for i := 1; i <= 5; i++ {
			if i == s.Cursor {
				b.WriteString(s.styles.Active.Render("  ^  "))
			} else {
				b.WriteString("      ")
			}
			b.WriteString(" ")
		}
		b.WriteString("\n")
	}

	return b.String()
}
