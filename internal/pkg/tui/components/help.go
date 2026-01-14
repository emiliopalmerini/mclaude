package components

import (
	"strings"

	"claude-watcher/internal/pkg/tui/theme"
)

// KeyBinding represents a key binding for the help bar
type KeyBinding struct {
	Key  string
	Desc string
}

// HelpBar renders a horizontal help bar with key bindings
type HelpBar struct {
	Bindings []KeyBinding
	styles   *theme.Styles
}

// NewHelpBar creates a new help bar
func NewHelpBar(bindings ...KeyBinding) HelpBar {
	return HelpBar{
		Bindings: bindings,
		styles:   theme.Default(),
	}
}

// SetBindings updates the key bindings
func (h *HelpBar) SetBindings(bindings ...KeyBinding) {
	h.Bindings = bindings
}

// View renders the help bar
func (h HelpBar) View() string {
	var parts []string
	for _, kb := range h.Bindings {
		parts = append(parts,
			h.styles.HelpKey.Render(kb.Key)+
				h.styles.Muted.Render(":"+kb.Desc))
	}
	return strings.Join(parts, " ")
}
