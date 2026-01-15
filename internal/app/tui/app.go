package tui

import (
	"claude-watcher/internal/analytics"
	analyticstui "claude-watcher/internal/analytics/inbound/tui"
	"claude-watcher/internal/pkg/tui/theme"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Screen identifies the current screen
type Screen int

const (
	ScreenOverview Screen = iota
	ScreenSessions
	ScreenDetail
	ScreenCosts
)

// App is the main dashboard TUI application
type App struct {
	service       *analytics.Service
	currentScreen Screen
	overview      *analyticstui.Overview
	sessions      *analyticstui.Sessions
	detail        *analyticstui.Detail
	costs         *analyticstui.Costs
	styles        *theme.Styles
	width         int
	height        int
}

// NewApp creates a new dashboard application
func NewApp(analyticsService *analytics.Service) *App {
	return &App{
		service:       analyticsService,
		currentScreen: ScreenOverview,
		overview:      analyticstui.NewOverview(analyticsService),
		sessions:      analyticstui.NewSessions(analyticsService),
		costs:         analyticstui.NewCosts(analyticsService),
		styles:        theme.Default(),
	}
}

// Init implements tea.Model
func (a *App) Init() tea.Cmd {
	return a.overview.Init()
}

// Update implements tea.Model
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "1":
			if a.currentScreen != ScreenOverview {
				a.currentScreen = ScreenOverview
				return a, a.overview.Init()
			}
		case "2":
			if a.currentScreen != ScreenSessions {
				a.currentScreen = ScreenSessions
				return a, a.sessions.Init()
			}
		case "3":
			if a.currentScreen != ScreenCosts {
				a.currentScreen = ScreenCosts
				return a, a.costs.Init()
			}
		case "esc":
			if a.currentScreen == ScreenDetail {
				a.currentScreen = ScreenSessions
				return a, nil
			}
		}

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height

	case analyticstui.SessionSelectedMsg:
		a.detail = analyticstui.NewDetail(a.service, msg.SessionID)
		a.currentScreen = ScreenDetail
		return a, a.detail.Init()
	}

	// Forward to current screen
	var cmd tea.Cmd
	switch a.currentScreen {
	case ScreenOverview:
		a.overview, cmd = a.overview.Update(msg)
	case ScreenSessions:
		a.sessions, cmd = a.sessions.Update(msg)
	case ScreenDetail:
		if a.detail != nil {
			a.detail, cmd = a.detail.Update(msg)
		}
	case ScreenCosts:
		a.costs, cmd = a.costs.Update(msg)
	}

	return a, cmd
}

// View implements tea.Model
func (a *App) View() string {
	header := a.renderHeader()
	nav := a.renderNav()

	// Separator line for clean visual hierarchy
	sep := lipgloss.NewStyle().
		Foreground(theme.Gray700).
		Render("────────────────────────────────────────────────────────────────")

	var content string
	switch a.currentScreen {
	case ScreenOverview:
		content = a.overview.View()
	case ScreenSessions:
		content = a.sessions.View()
	case ScreenDetail:
		if a.detail != nil {
			content = a.detail.View()
		}
	case ScreenCosts:
		content = a.costs.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, nav, sep, "", content)
}

func (a *App) renderHeader() string {
	// Bold, prominent title - typography as the focus
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.White).
		Render("CLAUDE WATCHER")

	// Subtle tagline
	tagline := lipgloss.NewStyle().
		Foreground(theme.Gray600).
		Render("Session Analytics")

	return lipgloss.JoinHorizontal(lipgloss.Bottom, title, "  ", tagline)
}

func (a *App) renderNav() string {
	items := []NavItem{
		{Key: "1", Label: "Overview", Active: a.currentScreen == ScreenOverview},
		{Key: "2", Label: "Sessions", Active: a.currentScreen == ScreenSessions || a.currentScreen == ScreenDetail},
		{Key: "3", Label: "Costs", Active: a.currentScreen == ScreenCosts},
	}
	nav := NewNavBar(items)
	return nav.View()
}
