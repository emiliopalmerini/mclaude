package tui

import (
	"context"
	"fmt"
	"time"

	"claude-watcher/internal/analytics"
	"claude-watcher/internal/pkg/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Overview displays aggregate metrics for the dashboard
type Overview struct {
	service *analytics.Service
	metrics analytics.OverviewMetrics
	loading bool
	err     error
	styles  *theme.Styles
	width   int
	height  int
}

// NewOverview creates a new overview screen
func NewOverview(service *analytics.Service) *Overview {
	return &Overview{
		service: service,
		loading: true,
		styles:  theme.Default(),
	}
}

// Init implements tea.Model
func (o *Overview) Init() tea.Cmd {
	return o.loadMetrics()
}

func (o *Overview) loadMetrics() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		metrics, err := o.service.GetOverview(ctx)
		if err != nil {
			return metricsErrorMsg{fmt.Errorf("load metrics: %w", err)}
		}
		return metricsLoadedMsg{metrics}
	}
}

// Update implements tea.Model
func (o *Overview) Update(msg tea.Msg) (*Overview, tea.Cmd) {
	switch msg := msg.(type) {
	case metricsLoadedMsg:
		o.loading = false
		o.metrics = msg.metrics
		return o, nil

	case metricsErrorMsg:
		o.loading = false
		o.err = msg.err
		return o, nil

	case tea.WindowSizeMsg:
		o.width = msg.Width
		o.height = msg.Height
		return o, nil

	case tea.KeyMsg:
		if msg.String() == "r" {
			o.loading = true
			o.err = nil
			return o, o.loadMetrics()
		}
	}

	return o, nil
}

// View implements tea.Model
func (o *Overview) View() string {
	if o.loading {
		return o.styles.Muted.Render("Loading metrics...")
	}

	if o.err != nil {
		return o.styles.Error.Render(fmt.Sprintf("Error: %v", o.err))
	}

	title := o.styles.Title.Render("Dashboard Overview (Last 7 Days)")

	cards := o.buildCards()
	cardView := RenderMetricCards(cards, o.width)

	help := o.styles.Help.Render("r: refresh  q: quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", cardView, "", help)
}

func (o *Overview) buildCards() []MetricCard {
	m := o.metrics

	return []MetricCard{
		{
			Title:    "Sessions",
			Value:    fmt.Sprintf("%d", m.TotalSessions),
			Subtitle: "Last 7 days",
		},
		{
			Title:    "Total Cost",
			Value:    fmt.Sprintf("$%.2f", m.TotalCost),
			Subtitle: "Estimated",
		},
		{
			Title:    "Tokens Used",
			Value:    formatTokens(m.Tokens.Total()),
			Subtitle: fmt.Sprintf("%s in / %s out", formatTokens(m.Tokens.Input), formatTokens(m.Tokens.Output)),
		},
		{
			Title:    "Limit Hits",
			Value:    fmt.Sprintf("%d", m.LimitHits),
			Subtitle: o.formatLastLimit(),
		},
	}
}

func (o *Overview) formatLastLimit() string {
	if o.metrics.LastLimitHit == nil {
		return "No limits hit"
	}
	return fmt.Sprintf("Last: %s", o.metrics.LastLimitHit.Format(time.RFC822))
}

func formatTokens(tokens int64) string {
	if tokens >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1_000_000)
	}
	if tokens >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(tokens)/1_000)
	}
	return fmt.Sprintf("%d", tokens)
}

// Message types
type metricsLoadedMsg struct {
	metrics analytics.OverviewMetrics
}

type metricsErrorMsg struct {
	err error
}
