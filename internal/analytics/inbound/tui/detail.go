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

// Detail displays session details
type Detail struct {
	service   *analytics.Service
	sessionID string
	session   analytics.SessionDetail
	loading   bool
	err       error
	styles    *theme.Styles
	width     int
	height    int
}

// NewDetail creates a new detail screen
func NewDetail(service *analytics.Service, sessionID string) *Detail {
	return &Detail{
		service:   service,
		sessionID: sessionID,
		loading:   true,
		styles:    theme.Default(),
	}
}

// Init implements tea.Model
func (d *Detail) Init() tea.Cmd {
	return d.loadSession()
}

func (d *Detail) loadSession() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		session, err := d.service.GetSession(ctx, d.sessionID)
		if err != nil {
			return detailErrorMsg{err}
		}
		return detailLoadedMsg{session}
	}
}

// Update implements tea.Model
func (d *Detail) Update(msg tea.Msg) (*Detail, tea.Cmd) {
	switch msg := msg.(type) {
	case detailLoadedMsg:
		d.loading = false
		d.session = msg.session
		return d, nil

	case detailErrorMsg:
		d.loading = false
		d.err = msg.err
		return d, nil

	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		return d, nil

	case tea.KeyMsg:
		if msg.String() == "r" {
			d.loading = true
			d.err = nil
			return d, d.loadSession()
		}
	}

	return d, nil
}

// View implements tea.Model
func (d *Detail) View() string {
	if d.loading {
		return d.styles.Muted.Render("Loading session...")
	}

	if d.err != nil {
		return d.styles.Error.Render(fmt.Sprintf("Error: %v", d.err))
	}

	s := d.session

	// Title
	title := d.styles.Title.Render("Session Detail")

	// Session ID
	idLine := d.renderField("ID", s.SessionID[:min(8, len(s.SessionID))])

	// Basic info section
	timeStr := s.Timestamp.Format("Jan 02, 2006 15:04:05")
	basicInfo := lipgloss.JoinVertical(lipgloss.Left,
		d.renderField("Time", timeStr),
		d.renderField("Duration", formatDuration(s.DurationSeconds)),
		d.renderField("Project", filepath.Base(s.WorkingDirectory)),
		d.renderField("Branch", orDash(s.GitBranch)),
		d.renderField("Model", orDash(s.Model)),
		d.renderField("Exit", orDash(s.ExitReason)),
	)

	// Metrics section
	metricsTitle := d.styles.Subtitle.Render("Metrics")
	metrics := lipgloss.JoinVertical(lipgloss.Left,
		d.renderField("Cost", fmt.Sprintf("$%.4f", s.EstimatedCost)),
		d.renderField("Prompts", fmt.Sprintf("%d", s.UserPrompts)),
		d.renderField("Responses", fmt.Sprintf("%d", s.AssistantResponses)),
		d.renderField("Tool Calls", fmt.Sprintf("%d", s.ToolCalls)),
		d.renderField("Errors", fmt.Sprintf("%d", s.ErrorsCount)),
	)

	// Tokens section
	tokensTitle := d.styles.Subtitle.Render("Tokens")
	tokens := lipgloss.JoinVertical(lipgloss.Left,
		d.renderField("Input", formatTokens(s.Tokens.Input)),
		d.renderField("Output", formatTokens(s.Tokens.Output)),
		d.renderField("Thinking", formatTokens(s.Tokens.Thinking)),
		d.renderField("Cache Read", formatTokens(s.Tokens.CacheRead)),
		d.renderField("Cache Write", formatTokens(s.Tokens.CacheWrite)),
	)

	// Quality section (if rated)
	var quality string
	if s.Rating != nil {
		qualityTitle := d.styles.Subtitle.Render("Quality Feedback")
		qualityContent := lipgloss.JoinVertical(lipgloss.Left,
			d.renderField("Rating", formatRating(s.Rating)),
			d.renderField("Prompt Spec", formatRating(s.PromptSpecificity)),
			d.renderField("Completion", formatRating(s.TaskCompletion)),
			d.renderField("Confidence", formatRating(s.CodeConfidence)),
		)
		if s.Notes != "" {
			qualityContent = lipgloss.JoinVertical(lipgloss.Left, qualityContent, d.renderField("Notes", truncate(s.Notes, 50)))
		}
		quality = lipgloss.JoinVertical(lipgloss.Left, "", qualityTitle, qualityContent)
	}

	// Summary
	var summary string
	if s.Summary != "" {
		summaryTitle := d.styles.Subtitle.Render("Summary")
		summaryText := d.styles.Body.Render(truncate(s.Summary, 200))
		summary = lipgloss.JoinVertical(lipgloss.Left, "", summaryTitle, summaryText)
	}

	help := d.styles.Help.Render("esc/2: back to sessions  1: overview  r: refresh  q: quit")

	content := lipgloss.JoinVertical(lipgloss.Left,
		title,
		idLine,
		"",
		basicInfo,
		"",
		metricsTitle,
		metrics,
		"",
		tokensTitle,
		tokens,
		quality,
		summary,
		"",
		help,
	)

	return content
}

func (d *Detail) renderField(label, value string) string {
	labelStyle := d.styles.BoldMuted.Copy().Width(14)
	valueStyle := d.styles.Body
	return lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render(label+":"),
		valueStyle.Render(value),
	)
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
}

func formatRating(r *int) string {
	if r == nil {
		return "-"
	}
	return fmt.Sprintf("%d/5", *r)
}

func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > maxLen {
		return s[:maxLen-3] + "..."
	}
	return s
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Message types
type detailLoadedMsg struct {
	session analytics.SessionDetail
}

type detailErrorMsg struct {
	err error
}
