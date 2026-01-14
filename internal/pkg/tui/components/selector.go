package components

import (
	"fmt"
	"strings"

	"claude-watcher/internal/pkg/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
)

// Option represents a selectable option
type Option struct {
	Label string
	Value string
}

// Selector is a multi-select list component
type Selector struct {
	Label    string
	Options  []Option
	Selected map[string]bool
	Cursor   int
	Focused  bool
	Multi    bool // Allow multiple selections
	styles   *theme.Styles
}

// NewSelector creates a new single-select selector
func NewSelector(label string, options []Option) Selector {
	return Selector{
		Label:    label,
		Options:  options,
		Selected: make(map[string]bool),
		Cursor:   0,
		Focused:  false,
		Multi:    false,
		styles:   theme.Default(),
	}
}

// NewMultiSelector creates a new multi-select selector
func NewMultiSelector(label string, options []Option) Selector {
	s := NewSelector(label, options)
	s.Multi = true
	return s
}

// Focus sets the selector as focused
func (s *Selector) Focus() {
	s.Focused = true
}

// Blur removes focus from the selector
func (s *Selector) Blur() {
	s.Focused = false
}

// SelectedValues returns the selected option values
func (s Selector) SelectedValues() []string {
	var result []string
	for _, opt := range s.Options {
		if s.Selected[opt.Value] {
			result = append(result, opt.Value)
		}
	}
	return result
}

// SetSelected sets the selected values
func (s *Selector) SetSelected(values []string) {
	s.Selected = make(map[string]bool)
	for _, v := range values {
		s.Selected[v] = true
	}
}

// Update handles key events for the selector
func (s Selector) Update(msg tea.Msg) (Selector, tea.Cmd) {
	if !s.Focused || len(s.Options) == 0 {
		return s, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "k", "up":
			if s.Cursor > 0 {
				s.Cursor--
			}
		case "j", "down":
			if s.Cursor < len(s.Options)-1 {
				s.Cursor++
			}
		case " ":
			s.toggle()
		case "enter":
			if !s.Multi {
				// Single select: select current and return
				s.Selected = make(map[string]bool)
				s.Selected[s.Options[s.Cursor].Value] = true
			}
		}
	}

	return s, nil
}

func (s *Selector) toggle() {
	if len(s.Options) == 0 {
		return
	}
	value := s.Options[s.Cursor].Value
	if s.Multi {
		s.Selected[value] = !s.Selected[value]
	} else {
		// Single select: only one can be selected
		s.Selected = make(map[string]bool)
		s.Selected[value] = true
	}
}

// View renders the selector
func (s Selector) View() string {
	var b strings.Builder

	b.WriteString(s.styles.Subtitle.Render(s.Label))
	b.WriteString("\n\n")

	for i, opt := range s.Options {
		isSelected := s.Selected[opt.Value]
		isCursor := s.Focused && i == s.Cursor

		var indicator string
		if isCursor {
			indicator = s.styles.Active.Render(">")
		} else {
			indicator = " "
		}

		var bullet string
		if isSelected {
			bullet = s.styles.Selected.Render("[x]")
		} else {
			bullet = s.styles.Muted.Render("[ ]")
		}

		var label string
		if isCursor {
			label = s.styles.Cursor.Render(opt.Label)
		} else if isSelected {
			label = s.styles.Selected.Render(opt.Label)
		} else {
			label = s.styles.Muted.Render(opt.Label)
		}

		b.WriteString(fmt.Sprintf("  %s %s %s\n", indicator, bullet, label))
	}

	return b.String()
}
