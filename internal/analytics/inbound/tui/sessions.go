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

const pageSize = 15

// Sessions displays a list of sessions
type Sessions struct {
	service  *analytics.Service
	sessions []analytics.SessionSummary
	total    int
	loading  bool
	err      error
	cursor   int
	page     int
	styles   *theme.Styles
	width    int
	height   int
}

// NewSessions creates a new sessions list screen
func NewSessions(service *analytics.Service) *Sessions {
	return &Sessions{
		service: service,
		loading: true,
		styles:  theme.Default(),
	}
}

// Init implements tea.Model
func (s *Sessions) Init() tea.Cmd {
	return s.loadSessions()
}

func (s *Sessions) loadSessions() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		filter := analytics.SessionFilter{
			Limit:  pageSize,
			Offset: s.page * pageSize,
		}
		sessions, total, err := s.service.ListSessions(ctx, filter)
		if err != nil {
			return sessionsErrorMsg{err}
		}
		return sessionsLoadedMsg{sessions, total}
	}
}

// Update implements tea.Model
func (s *Sessions) Update(msg tea.Msg) (*Sessions, tea.Cmd) {
	switch msg := msg.(type) {
	case sessionsLoadedMsg:
		s.loading = false
		s.sessions = msg.sessions
		s.total = msg.total
		s.cursor = 0
		return s, nil

	case sessionsErrorMsg:
		s.loading = false
		s.err = msg.err
		return s, nil

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		return s, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if s.cursor < len(s.sessions)-1 {
				s.cursor++
			}
		case "k", "up":
			if s.cursor > 0 {
				s.cursor--
			}
		case "n", "pgdown":
			if (s.page+1)*pageSize < s.total {
				s.page++
				s.loading = true
				return s, s.loadSessions()
			}
		case "p", "pgup":
			if s.page > 0 {
				s.page--
				s.loading = true
				return s, s.loadSessions()
			}
		case "r":
			s.loading = true
			s.err = nil
			return s, s.loadSessions()
		case "enter":
			if len(s.sessions) > 0 {
				return s, func() tea.Msg {
					return SessionSelectedMsg{SessionID: s.sessions[s.cursor].SessionID}
				}
			}
		}
	}

	return s, nil
}

// View implements tea.Model
func (s *Sessions) View() string {
	if s.loading {
		return s.styles.Muted.Render("Loading sessions...")
	}

	if s.err != nil {
		return s.styles.Error.Render(fmt.Sprintf("Error: %v", s.err))
	}

	title := s.styles.Title.Render("Sessions")

	// Pagination info
	start := s.page*pageSize + 1
	end := start + len(s.sessions) - 1
	pageInfo := s.styles.Muted.Render(fmt.Sprintf("Showing %d-%d of %d", start, end, s.total))

	// Table header
	header := s.renderHeader()

	// Table rows
	var rows []string
	for i, session := range s.sessions {
		rows = append(rows, s.renderRow(session, i == s.cursor))
	}

	table := lipgloss.JoinVertical(lipgloss.Left, append([]string{header}, rows...)...)

	help := s.styles.Help.Render("j/k: navigate  n/p: page  enter: select  r: refresh  1: overview  q: quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, pageInfo, "", table, "", help)
}

func (s *Sessions) renderHeader() string {
	headerStyle := lipgloss.NewStyle().
		Foreground(theme.Gray500).
		Bold(true)

	cols := []string{
		headerStyle.Copy().Width(12).Render("TIME"),
		headerStyle.Copy().Width(25).Render("PROJECT"),
		headerStyle.Copy().Width(15).Render("MODEL"),
		headerStyle.Copy().Width(10).Render("COST"),
		headerStyle.Copy().Width(8).Render("TOOLS"),
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

func (s *Sessions) renderRow(session analytics.SessionSummary, selected bool) string {
	// Format timestamp
	timeStr := session.Timestamp.Format("Jan 02 15:04")

	// Truncate working dir to just the project name
	project := filepath.Base(session.WorkingDir)
	if project == "." || project == "" {
		project = "-"
	}
	if len(project) > 23 {
		project = project[:20] + "..."
	}

	// Truncate model name
	model := session.Model
	if model == "" {
		model = "-"
	}
	// Remove common prefixes for brevity
	model = strings.TrimPrefix(model, "claude-")
	if len(model) > 13 {
		model = model[:10] + "..."
	}

	var style lipgloss.Style
	if selected {
		// Inverted selection: white text on dark background
		style = lipgloss.NewStyle().
			Foreground(theme.Black).
			Background(theme.White).
			Bold(true)
	} else {
		style = lipgloss.NewStyle().
			Foreground(theme.Gray400)
	}

	cols := []string{
		style.Copy().Width(12).Render(timeStr),
		style.Copy().Width(25).Render(project),
		style.Copy().Width(15).Render(model),
		style.Copy().Width(10).Render(fmt.Sprintf("$%.2f", session.EstimatedCost)),
		style.Copy().Width(8).Render(fmt.Sprintf("%d", session.ToolCalls)),
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

// SelectedSessionID returns the currently selected session ID
func (s *Sessions) SelectedSessionID() string {
	if s.cursor >= 0 && s.cursor < len(s.sessions) {
		return s.sessions[s.cursor].SessionID
	}
	return ""
}

// Message types
type sessionsLoadedMsg struct {
	sessions []analytics.SessionSummary
	total    int
}

type sessionsErrorMsg struct {
	err error
}

// SessionSelectedMsg is sent when a session is selected
type SessionSelectedMsg struct {
	SessionID string
}
