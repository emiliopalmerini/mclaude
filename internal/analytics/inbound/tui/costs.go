package tui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"claude-watcher/internal/analytics"
	"claude-watcher/internal/pkg/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Costs displays cost analysis breakdown
type Costs struct {
	service *analytics.Service
	data    analytics.CostsBreakdown
	loading bool
	err     error
	styles  *theme.Styles
	width   int
	height  int
}

// NewCosts creates a new costs breakdown screen
func NewCosts(service *analytics.Service) *Costs {
	return &Costs{
		service: service,
		loading: true,
		styles:  theme.Default(),
	}
}

// Init implements tea.Model
func (c *Costs) Init() tea.Cmd {
	return c.loadCosts()
}

func (c *Costs) loadCosts() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		data, err := c.service.GetCostsBreakdown(ctx)
		if err != nil {
			return costsErrorMsg{fmt.Errorf("load costs: %w", err)}
		}
		return costsLoadedMsg{data}
	}
}

// Update implements tea.Model
func (c *Costs) Update(msg tea.Msg) (*Costs, tea.Cmd) {
	switch msg := msg.(type) {
	case costsLoadedMsg:
		c.loading = false
		c.data = msg.data
		return c, nil

	case costsErrorMsg:
		c.loading = false
		c.err = msg.err
		return c, nil

	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height
		return c, nil

	case tea.KeyMsg:
		if msg.String() == "r" {
			c.loading = true
			c.err = nil
			return c, c.loadCosts()
		}
	}

	return c, nil
}

// View implements tea.Model
func (c *Costs) View() string {
	if c.loading {
		return c.styles.Muted.Render("Loading costs...")
	}

	if c.err != nil {
		return c.styles.Error.Render(fmt.Sprintf("Error: %v", c.err))
	}

	title := c.styles.Title.Render("Cost Analysis")

	// Summary cards
	cards := c.buildSummaryCards()
	cardView := RenderMetricCards(cards, c.width)

	// Daily trend with sparkline
	trendView := c.buildTrendView()

	// Model breakdown table
	modelView := c.buildModelTable()

	// Project breakdown table
	projectView := c.buildProjectTable()

	help := c.styles.Help.Render("r: refresh  q: quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		title, "",
		cardView, "",
		trendView, "",
		modelView, "",
		projectView, "",
		help,
	)
}

func (c *Costs) buildSummaryCards() []MetricCard {
	return []MetricCard{
		{
			Title:    "Total Cost",
			Value:    fmt.Sprintf("$%.2f", c.data.TotalCost),
			Subtitle: "All time",
		},
		{
			Title:    "Today",
			Value:    fmt.Sprintf("$%.2f", c.data.TodayCost),
			Subtitle: "Today's spend",
		},
		{
			Title:    "This Week",
			Value:    fmt.Sprintf("$%.2f", c.data.WeekCost),
			Subtitle: "Last 7 days",
		},
	}
}

func (c *Costs) buildTrendView() string {
	if len(c.data.DailyTrend) == 0 {
		return c.styles.Muted.Render("No daily data available")
	}

	// Extract values for sparkline
	values := make([]float64, len(c.data.DailyTrend))
	for i, d := range c.data.DailyTrend {
		values[i] = d.Cost
	}

	sparkline := RenderSparkline(values)

	// Build date labels (first and last)
	dateRange := ""
	if len(c.data.DailyTrend) > 0 {
		first := c.data.DailyTrend[0].Date
		last := c.data.DailyTrend[len(c.data.DailyTrend)-1].Date
		dateRange = fmt.Sprintf("%s → %s", first, last)
	}

	header := c.styles.Subtitle.Render("7-DAY TREND")
	// Sparkline rendered in white for prominence
	chart := lipgloss.NewStyle().
		Foreground(theme.White).
		Render(sparkline)
	dates := c.styles.Muted.Render(dateRange)

	return lipgloss.JoinVertical(lipgloss.Left, header, chart, dates)
}

func (c *Costs) buildModelTable() string {
	if len(c.data.ByModel) == 0 {
		return c.styles.Muted.Render("No model data available")
	}

	header := c.styles.Subtitle.Render("COST BY MODEL")
	subheader := c.styles.Muted.Render("Last 30 days")

	// Table header - uppercase for clean hierarchy
	headerStyle := lipgloss.NewStyle().
		Foreground(theme.Gray500).
		Bold(true)
	tableHeader := fmt.Sprintf("%-35s %8s %10s %12s",
		"MODEL", "SESSIONS", "COST", "$/M TOKENS")
	headerLine := headerStyle.Render(tableHeader)

	// Table rows
	rowStyle := lipgloss.NewStyle().Foreground(theme.Gray400)
	var rows []string
	for _, m := range c.data.ByModel {
		model := truncateString(m.Model, 35)
		row := fmt.Sprintf("%-35s %8d %10s %12.2f",
			model,
			m.Sessions,
			fmt.Sprintf("$%.2f", m.Cost),
			m.CostPerMillionToks,
		)
		rows = append(rows, rowStyle.Render(row))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		subheader,
		"",
		headerLine,
		strings.Join(rows, "\n"),
	)
}

func (c *Costs) buildProjectTable() string {
	if len(c.data.ByProject) == 0 {
		return c.styles.Muted.Render("No project data available")
	}

	header := c.styles.Subtitle.Render("TOP PROJECTS")
	subheader := c.styles.Muted.Render("Last 30 days by cost")

	// Table header - uppercase for clean hierarchy
	headerStyle := lipgloss.NewStyle().
		Foreground(theme.Gray500).
		Bold(true)
	tableHeader := fmt.Sprintf("%-45s %8s %10s",
		"PROJECT", "SESSIONS", "COST")
	headerLine := headerStyle.Render(tableHeader)

	// Table rows
	rowStyle := lipgloss.NewStyle().Foreground(theme.Gray400)
	var rows []string
	for _, p := range c.data.ByProject {
		// Show just the last component of the path
		project := filepath.Base(p.Project)
		if project == "." || project == "" {
			project = p.Project
		}
		project = truncateString(project, 45)

		row := fmt.Sprintf("%-45s %8d %10s",
			project,
			p.Sessions,
			fmt.Sprintf("$%.2f", p.Cost),
		)
		rows = append(rows, rowStyle.Render(row))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		subheader,
		"",
		headerLine,
		strings.Join(rows, "\n"),
	)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// Message types
type costsLoadedMsg struct {
	data analytics.CostsBreakdown
}

type costsErrorMsg struct {
	err error
}
