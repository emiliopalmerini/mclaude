package theme

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// Styles contains all shared TUI styles
type Styles struct {
	// Text styles
	Title       lipgloss.Style
	Subtitle    lipgloss.Style
	Body        lipgloss.Style
	Muted       lipgloss.Style
	Bold        lipgloss.Style
	BoldMuted   lipgloss.Style
	Highlighted lipgloss.Style

	// Interactive elements
	Cursor     lipgloss.Style
	Selected   lipgloss.Style
	Unselected lipgloss.Style
	Active     lipgloss.Style
	Inactive   lipgloss.Style

	// Help and hints
	Help    lipgloss.Style
	HelpKey lipgloss.Style

	// Layout
	Container lipgloss.Style
	Card      lipgloss.Style
	Border    lipgloss.Style

	// Progress indicators
	ProgressActive   lipgloss.Style
	ProgressInactive lipgloss.Style

	// Status indicators
	Success lipgloss.Style
	Warning lipgloss.Style
	Error   lipgloss.Style
	Info    lipgloss.Style
}

var (
	defaultStyles *Styles
	once          sync.Once
)

// Default returns the singleton default Styles instance
func Default() *Styles {
	once.Do(func() {
		defaultStyles = newStyles()
	})
	return defaultStyles
}

func newStyles() *Styles {
	return &Styles{
		// Text styles
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(White).
			MarginBottom(1),

		Subtitle: lipgloss.NewStyle().
			Foreground(Purple).
			Bold(true).
			MarginBottom(1),

		Body: lipgloss.NewStyle().
			Foreground(LightGray),

		Muted: lipgloss.NewStyle().
			Foreground(DimGray),

		Bold: lipgloss.NewStyle().
			Bold(true).
			Foreground(White),

		BoldMuted: lipgloss.NewStyle().
			Bold(true).
			Foreground(DimGray),

		Highlighted: lipgloss.NewStyle().
			Foreground(BrightPurple).
			Bold(true),

		// Interactive elements
		Cursor: lipgloss.NewStyle().
			Foreground(BrightPurple).
			Bold(true),

		Selected: lipgloss.NewStyle().
			Foreground(Purple),

		Unselected: lipgloss.NewStyle().
			Foreground(DimGray),

		Active: lipgloss.NewStyle().
			Foreground(BrightPurple).
			Bold(true),

		Inactive: lipgloss.NewStyle().
			Foreground(DimGray),

		// Help and hints
		Help: lipgloss.NewStyle().
			Foreground(DimGray).
			MarginTop(1),

		HelpKey: lipgloss.NewStyle().
			Foreground(LightGray).
			Bold(true),

		// Layout
		Container: lipgloss.NewStyle().
			Padding(1, 2),

		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DarkGray).
			Padding(1, 2),

		Border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(DimGray),

		// Progress indicators
		ProgressActive: lipgloss.NewStyle().
			Foreground(Purple),

		ProgressInactive: lipgloss.NewStyle().
			Foreground(DimGray),

		// Status indicators
		Success: lipgloss.NewStyle().
			Foreground(Success),

		Warning: lipgloss.NewStyle().
			Foreground(Warning),

		Error: lipgloss.NewStyle().
			Foreground(Error),

		Info: lipgloss.NewStyle().
			Foreground(Info),
	}
}
