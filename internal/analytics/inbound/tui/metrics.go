package tui

import (
	"claude-watcher/internal/pkg/tui/theme"

	"github.com/charmbracelet/lipgloss"
)

// MetricCard displays a single KPI metric
type MetricCard struct {
	Title    string
	Value    string
	Subtitle string
}

// View renders the metric card with minimal, typography-focused design
func (m MetricCard) View(width int) string {
	styles := theme.Default()

	card := styles.Card.Copy().Width(width)

	// Large, bold value as focal point
	value := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.White).
		Render(m.Value)

	// Title above value, uppercase for hierarchy
	title := lipgloss.NewStyle().
		Foreground(theme.Gray500).
		Render(m.Title)

	// Subtitle below, muted
	subtitle := styles.Muted.Render(m.Subtitle)

	content := lipgloss.JoinVertical(lipgloss.Left, title, value, subtitle)
	return card.Render(content)
}

// RenderMetricCards renders a grid of metric cards
func RenderMetricCards(cards []MetricCard, totalWidth int) string {
	if len(cards) == 0 {
		return ""
	}

	// Default width if not set yet
	if totalWidth <= 0 {
		totalWidth = 80
	}

	// Calculate card width (2 cards per row with spacing)
	cardWidth := (totalWidth - 4) / 2
	if cardWidth < 20 {
		cardWidth = 20
	}

	var rows []string

	// Render cards in pairs
	for i := 0; i < len(cards); i += 2 {
		var rowCards []string
		rowCards = append(rowCards, cards[i].View(cardWidth))

		if i+1 < len(cards) {
			rowCards = append(rowCards, cards[i+1].View(cardWidth))
		}

		row := lipgloss.JoinHorizontal(lipgloss.Top, rowCards...)
		rows = append(rows, row)
	}

	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
