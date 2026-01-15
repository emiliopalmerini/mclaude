package tui

import (
	"strings"

	"claude-watcher/internal/pkg/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// NavItem represents a navigation item
type NavItem struct {
	Key    string
	Label  string
	Active bool
}

// NavBar renders a navigation bar
type NavBar struct {
	Items  []NavItem
	styles *theme.Styles
}

// NewNavBar creates a new navigation bar
func NewNavBar(items []NavItem) *NavBar {
	return &NavBar{
		Items:  items,
		styles: theme.Default(),
	}
}

// View renders the navigation bar as toggle-style tabs
func (n NavBar) View() string {
	var items []string

	for _, item := range n.Items {
		var rendered string
		if item.Active {
			// Active tab: inverted colors (white on black effect)
			rendered = n.styles.Active.Render(item.Label)
		} else {
			// Inactive: muted with key hint
			key := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#525252")).
				Render("[" + item.Key + "]")
			label := n.styles.Inactive.Render(item.Label)
			rendered = key + " " + label
		}
		items = append(items, rendered)
	}

	// Join with generous spacing for clean look
	sep := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#404040")).
		Render("  /  ")

	return strings.Join(items, sep)
}
