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
		// Text styles - typography-focused, high contrast
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(White).
			MarginBottom(1),

		Subtitle: lipgloss.NewStyle().
			Foreground(Gray300).
			Bold(true).
			MarginTop(1),

		Body: lipgloss.NewStyle().
			Foreground(Gray400),

		Muted: lipgloss.NewStyle().
			Foreground(Gray500),

		Bold: lipgloss.NewStyle().
			Bold(true).
			Foreground(White),

		BoldMuted: lipgloss.NewStyle().
			Bold(true).
			Foreground(Gray500),

		Highlighted: lipgloss.NewStyle().
			Foreground(White).
			Bold(true),

		// Interactive elements - inverted for selection
		Cursor: lipgloss.NewStyle().
			Foreground(Black).
			Background(White).
			Bold(true),

		Selected: lipgloss.NewStyle().
			Foreground(White).
			Bold(true),

		Unselected: lipgloss.NewStyle().
			Foreground(Gray500),

		Active: lipgloss.NewStyle().
			Foreground(Black).
			Background(White).
			Bold(true).
			Padding(0, 1),

		Inactive: lipgloss.NewStyle().
			Foreground(Gray500).
			Padding(0, 1),

		// Help and hints
		Help: lipgloss.NewStyle().
			Foreground(Gray600).
			MarginTop(2),

		HelpKey: lipgloss.NewStyle().
			Foreground(Gray400).
			Bold(true),

		// Layout - minimal borders, clean lines
		Container: lipgloss.NewStyle().
			Padding(1, 2),

		Card: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(Gray700).
			Padding(0, 2).
			MarginRight(2),

		Border: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(Gray700),

		// Progress indicators
		ProgressActive: lipgloss.NewStyle().
			Foreground(White),

		ProgressInactive: lipgloss.NewStyle().
			Foreground(Gray700),

		// Status indicators - subtle differentiation
		Success: lipgloss.NewStyle().
			Foreground(White),

		Warning: lipgloss.NewStyle().
			Foreground(Gray300),

		Error: lipgloss.NewStyle().
			Foreground(Gray400).
			Italic(true),

		Info: lipgloss.NewStyle().
			Foreground(Gray400),
	}
}
