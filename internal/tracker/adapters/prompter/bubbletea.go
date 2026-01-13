package prompter

import (
	"claude-watcher/internal/tracker/domain"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// BubbleTeaPrompter collects quality feedback using a TUI
type BubbleTeaPrompter struct {
	logger domain.Logger
}

// NewBubbleTeaPrompter creates a new Bubbletea prompter
func NewBubbleTeaPrompter(logger domain.Logger) *BubbleTeaPrompter {
	return &BubbleTeaPrompter{logger: logger}
}

// CollectQualityData prompts the user for session feedback via TUI.
// Returns empty QualityData if TTY is unavailable or user skips.
func (p *BubbleTeaPrompter) CollectQualityData(tags []domain.Tag) (domain.QualityData, error) {
	if !isTerminal() {
		p.logger.Debug("TTY not available, skipping quality prompts")
		return domain.QualityData{}, nil
	}

	m := newModel(tags)
	prog := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := prog.Run()
	if err != nil {
		return domain.QualityData{}, err
	}

	result := finalModel.(model)
	if result.cancelled {
		return domain.QualityData{}, nil
	}

	return result.toQualityData(), nil
}

func isTerminal() bool {
	_, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	return err == nil
}

// Steps in the wizard
const (
	stepTagsTaskType = iota
	stepPromptSpecificity
	stepTaskCompletion
	stepCodeConfidence
	stepRating
	stepNotes
	stepDone
)

// Scale definitions with labels and descriptions
type scaleInfo struct {
	label    string
	lowDesc  string
	highDesc string
}

var scaleLabels = map[int]scaleInfo{
	stepPromptSpecificity: {"Prompt Specificity", "Minimal/vague", "Highly detailed"},
	stepTaskCompletion:    {"Task Completion", "Abandoned", "Fully completed"},
	stepCodeConfidence:    {"Code Confidence", "Very uncertain", "Highly confident"},
	stepRating:            {"Session Satisfaction", "Poor", "Excellent"},
}

// Styles inspired by Claude Code
type styles struct {
	title       lipgloss.Style
	subtitle    lipgloss.Style
	cursor      lipgloss.Style
	selected    lipgloss.Style
	unselected  lipgloss.Style
	help        lipgloss.Style
	helpKey     lipgloss.Style
	container   lipgloss.Style
	indicator   lipgloss.Style
	numberHint  lipgloss.Style
	activeNum   lipgloss.Style
	progressBar lipgloss.Style
	progressDot lipgloss.Style
}

func newStyles() styles {
	purple := lipgloss.Color("#A855F7")
	brightPurple := lipgloss.Color("#C084FC")
	dimGray := lipgloss.Color("#6B7280")
	lightGray := lipgloss.Color("#9CA3AF")

	return styles{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			MarginBottom(1),
		subtitle: lipgloss.NewStyle().
			Foreground(purple).
			Bold(true).
			MarginBottom(1),
		cursor: lipgloss.NewStyle().
			Foreground(brightPurple).
			Bold(true),
		selected: lipgloss.NewStyle().
			Foreground(purple),
		unselected: lipgloss.NewStyle().
			Foreground(dimGray),
		help: lipgloss.NewStyle().
			Foreground(dimGray).
			MarginTop(1),
		helpKey: lipgloss.NewStyle().
			Foreground(lightGray).
			Bold(true),
		container: lipgloss.NewStyle().
			Padding(1, 2),
		indicator: lipgloss.NewStyle().
			Foreground(brightPurple).
			Bold(true),
		numberHint: lipgloss.NewStyle().
			Foreground(dimGray),
		activeNum: lipgloss.NewStyle().
			Foreground(brightPurple).
			Bold(true),
		progressBar: lipgloss.NewStyle().
			Foreground(dimGray),
		progressDot: lipgloss.NewStyle().
			Foreground(purple),
	}
}

// Vim modes for textarea
const (
	modeNormal = iota
	modeInsert
)

// Model for the TUI
type model struct {
	step int

	// Task type tags (only category remaining)
	taskTypeTags []domain.Tag
	selectedTags map[string]bool
	tagCursor    int

	// Scales (1-5, 0 = not set)
	promptSpecificity int
	taskCompletion    int
	codeConfidence    int
	rating            int
	scaleCursor       int // Current cursor position for scale steps

	// Notes textarea
	notesInput textarea.Model
	vimMode    int // modeNormal or modeInsert

	// State
	cancelled bool
	styles    styles
	width     int
	height    int
}

func newModel(tags []domain.Tag) model {
	// Filter to only task_type tags
	var taskTypeTags []domain.Tag
	for _, tag := range tags {
		if tag.Category == "task_type" {
			taskTypeTags = append(taskTypeTags, tag)
		}
	}

	// Start at tags if available, otherwise first scale
	startStep := stepPromptSpecificity
	if len(taskTypeTags) > 0 {
		startStep = stepTagsTaskType
	}

	ta := textarea.New()
	ta.Placeholder = "Any notes about this session..."
	ta.ShowLineNumbers = false
	ta.SetWidth(50)
	ta.SetHeight(3)
	ta.CharLimit = 500

	return model{
		step:              startStep,
		taskTypeTags:      taskTypeTags,
		selectedTags:      make(map[string]bool),
		tagCursor:         0,
		promptSpecificity: 0,
		taskCompletion:    0,
		codeConfidence:    0,
		rating:            0,
		scaleCursor:       3, // Default to middle
		notesInput:        ta,
		styles:            newStyles(),
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.step == stepNotes {
			return m.handleNotesKey(msg)
		}
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.notesInput.SetWidth(min(50, msg.Width-10))
		return m, nil
	}

	if m.step == stepNotes {
		var cmd tea.Cmd
		m.notesInput, cmd = m.notesInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) handleNotesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Always handle these
	if key == "ctrl+c" {
		m.cancelled = true
		return m, tea.Quit
	}

	if m.vimMode == modeInsert {
		// Insert mode - pass to textarea, esc exits
		switch key {
		case "esc":
			m.vimMode = modeNormal
			m.notesInput.Blur()
			return m, nil
		default:
			var cmd tea.Cmd
			m.notesInput, cmd = m.notesInput.Update(msg)
			return m, cmd
		}
	}

	// Normal mode
	switch key {
	case "i", "a":
		m.vimMode = modeInsert
		m.notesInput.Focus()
		return m, textarea.Blink
	case "q":
		m.cancelled = true
		return m, tea.Quit
	case "h", "backspace":
		return m.prevStep()
	case "l", "enter":
		m.step = stepDone
		return m, tea.Quit
	}

	return m, nil
}

func (m model) isScaleStep() bool {
	return m.step >= stepPromptSpecificity && m.step <= stepRating
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "ctrl+c", "q":
		m.cancelled = true
		return m, tea.Quit

	case "esc":
		m.step = stepDone
		return m, tea.Quit

	// Navigation: j/k for up/down (tags), h/l for prev/next step or scale value
	case "k", "up":
		if m.step == stepTagsTaskType {
			m.moveCursorUp()
		}
		return m, nil

	case "j", "down":
		if m.step == stepTagsTaskType {
			m.moveCursorDown()
		}
		return m, nil

	case "h", "left", "backspace":
		if m.isScaleStep() {
			if m.scaleCursor > 1 {
				m.scaleCursor--
			}
		} else {
			return m.prevStep()
		}
		return m, nil

	case "l", "right", "tab":
		if m.isScaleStep() {
			if m.scaleCursor < 5 {
				m.scaleCursor++
			}
		} else {
			return m.nextStep()
		}
		return m, nil

	case "enter":
		return m.nextStep()

	case " ":
		if m.step == stepTagsTaskType {
			m.toggleTag()
		} else if m.isScaleStep() {
			m.setCurrentScale(m.scaleCursor)
		}
		return m, nil

	case "1", "2", "3", "4", "5":
		if m.isScaleStep() {
			num := int(key[0] - '0')
			m.scaleCursor = num
			m.setCurrentScale(num)
		}
		return m, nil
	}

	return m, nil
}

func (m *model) setCurrentScale(value int) {
	switch m.step {
	case stepPromptSpecificity:
		m.promptSpecificity = value
	case stepTaskCompletion:
		m.taskCompletion = value
	case stepCodeConfidence:
		m.codeConfidence = value
	case stepRating:
		m.rating = value
	}
}

func (m *model) getCurrentScaleValue() int {
	switch m.step {
	case stepPromptSpecificity:
		return m.promptSpecificity
	case stepTaskCompletion:
		return m.taskCompletion
	case stepCodeConfidence:
		return m.codeConfidence
	case stepRating:
		return m.rating
	}
	return 0
}

func (m *model) nextStep() (tea.Model, tea.Cmd) {
	// Auto-set scale value if moving forward without explicit selection
	if m.isScaleStep() && m.getCurrentScaleValue() == 0 {
		m.setCurrentScale(m.scaleCursor)
	}

	m.step++
	if m.step >= stepDone {
		return m, tea.Quit
	}

	// Skip tags step if no tags available
	if m.step == stepTagsTaskType && len(m.taskTypeTags) == 0 {
		m.step = stepPromptSpecificity
	}

	// Reset cursor for new step
	m.tagCursor = 0
	m.scaleCursor = 3 // Reset to middle for scales

	if m.step == stepNotes {
		m.vimMode = modeNormal
		m.notesInput.Blur()
	}

	return m, nil
}

func (m *model) prevStep() (tea.Model, tea.Cmd) {
	m.step--

	// Skip tags step if no tags available
	if m.step == stepTagsTaskType && len(m.taskTypeTags) == 0 {
		m.step-- // Go before tags
	}

	if m.step < 0 {
		m.step = m.findFirstStep()
	}

	m.tagCursor = 0
	m.scaleCursor = 3

	return m, nil
}

func (m *model) findFirstStep() int {
	if len(m.taskTypeTags) > 0 {
		return stepTagsTaskType
	}
	return stepPromptSpecificity
}

func (m *model) moveCursorUp() {
	if len(m.taskTypeTags) > 0 && m.tagCursor > 0 {
		m.tagCursor--
	}
}

func (m *model) moveCursorDown() {
	if len(m.taskTypeTags) > 0 && m.tagCursor < len(m.taskTypeTags)-1 {
		m.tagCursor++
	}
}

func (m *model) toggleTag() {
	if len(m.taskTypeTags) == 0 {
		return
	}
	tagName := m.taskTypeTags[m.tagCursor].Name
	m.selectedTags[tagName] = !m.selectedTags[tagName]
}

func (m model) View() string {
	if m.step == stepDone {
		return ""
	}

	var b strings.Builder

	b.WriteString(m.styles.title.Render("Session Quality Feedback"))
	b.WriteString("  ")
	b.WriteString(m.renderProgress())
	b.WriteString("\n\n")

	switch m.step {
	case stepTagsTaskType:
		b.WriteString(m.viewTags())
	case stepPromptSpecificity, stepTaskCompletion, stepCodeConfidence, stepRating:
		b.WriteString(m.viewScale())
	case stepNotes:
		b.WriteString(m.viewNotes())
	}

	b.WriteString("\n")
	b.WriteString(m.renderHelp())

	return m.styles.container.Render(b.String())
}

func (m model) renderProgress() string {
	totalSteps := m.countTotalSteps()
	currentStep := m.countCurrentStep()

	var dots strings.Builder
	for i := 0; i < totalSteps; i++ {
		if i < currentStep {
			dots.WriteString(m.styles.progressDot.Render("●"))
		} else if i == currentStep {
			dots.WriteString(m.styles.progressDot.Render("◉"))
		} else {
			dots.WriteString(m.styles.progressBar.Render("○"))
		}
		if i < totalSteps-1 {
			dots.WriteString(" ")
		}
	}
	return dots.String()
}

func (m model) countTotalSteps() int {
	// 4 scales + notes = 5, plus tags if available = 6
	count := 5
	if len(m.taskTypeTags) > 0 {
		count++
	}
	return count
}

func (m model) countCurrentStep() int {
	offset := 0
	if len(m.taskTypeTags) > 0 {
		if m.step == stepTagsTaskType {
			return 0
		}
		offset = 1
	}

	switch m.step {
	case stepPromptSpecificity:
		return offset
	case stepTaskCompletion:
		return offset + 1
	case stepCodeConfidence:
		return offset + 2
	case stepRating:
		return offset + 3
	case stepNotes:
		return offset + 4
	}
	return 0
}

func (m model) viewTags() string {
	var b strings.Builder

	b.WriteString(m.styles.subtitle.Render("Task Type"))
	b.WriteString("\n\n")

	for i, tag := range m.taskTypeTags {
		isSelected := m.selectedTags[tag.Name]
		isCursor := i == m.tagCursor

		var indicator string
		if isCursor {
			indicator = m.styles.indicator.Render("❯")
		} else {
			indicator = " "
		}

		var bullet string
		if isSelected {
			bullet = m.styles.selected.Render("●")
		} else {
			bullet = m.styles.unselected.Render("○")
		}

		var name string
		if isCursor {
			name = m.styles.cursor.Render(tag.Name)
		} else if isSelected {
			name = m.styles.selected.Render(tag.Name)
		} else {
			name = m.styles.unselected.Render(tag.Name)
		}

		b.WriteString(fmt.Sprintf("  %s %s %s\n", indicator, bullet, name))
	}

	return b.String()
}

func (m model) viewScale() string {
	var b strings.Builder

	info := scaleLabels[m.step]
	currentValue := m.getCurrentScaleValue()

	b.WriteString(m.styles.subtitle.Render(info.label))
	b.WriteString("\n\n")

	// Scale labels
	b.WriteString("  ")
	b.WriteString(m.styles.unselected.Render(info.lowDesc))
	b.WriteString("                    ")
	b.WriteString(m.styles.unselected.Render(info.highDesc))
	b.WriteString("\n")

	// Number row
	b.WriteString("  ")
	for i := 1; i <= 5; i++ {
		isCursor := i == m.scaleCursor
		isSelected := i == currentValue
		numStr := fmt.Sprintf("%d", i)

		if isCursor {
			if isSelected {
				b.WriteString(m.styles.activeNum.Render("【" + numStr + "】"))
			} else {
				b.WriteString(m.styles.activeNum.Render("[ " + numStr + " ]"))
			}
		} else if isSelected {
			b.WriteString(m.styles.selected.Render("  " + numStr + "  "))
		} else {
			b.WriteString(m.styles.numberHint.Render("  " + numStr + "  "))
		}
		b.WriteString(" ")
	}
	b.WriteString("\n")

	// Cursor indicator row
	b.WriteString("  ")
	for i := 1; i <= 5; i++ {
		if i == m.scaleCursor {
			b.WriteString(m.styles.indicator.Render("  ▲  "))
		} else {
			b.WriteString("      ")
		}
		b.WriteString(" ")
	}
	b.WriteString("\n")

	return b.String()
}

func (m model) viewNotes() string {
	var b strings.Builder

	b.WriteString(m.styles.subtitle.Render("Notes (optional)"))

	// Vim mode indicator
	if m.vimMode == modeInsert {
		b.WriteString("  ")
		b.WriteString(m.styles.activeNum.Render("-- INSERT --"))
	} else {
		b.WriteString("  ")
		b.WriteString(m.styles.unselected.Render("-- NORMAL --"))
	}
	b.WriteString("\n\n")

	b.WriteString(m.notesInput.View())
	b.WriteString("\n")

	return b.String()
}

type keyBinding struct {
	key  string
	desc string
}

func (m model) getKeyBindings() []keyBinding {
	switch {
	case m.step == stepTagsTaskType:
		return []keyBinding{
			{"j/k", "nav"},
			{"spc", "sel"},
			{"⏎", "→"},
			{"q", "quit"},
		}
	case m.isScaleStep():
		return []keyBinding{
			{"h/l", "mov"},
			{"1-5", "sel"},
			{"spc", "ok"},
			{"⏎", "→"},
			{"q", "quit"},
		}
	case m.step == stepNotes:
		if m.vimMode == modeInsert {
			return []keyBinding{
				{"esc", "normal"},
			}
		}
		return []keyBinding{
			{"i", "ins"},
			{"h", "←"},
			{"l/⏎", "done"},
			{"q", "quit"},
		}
	}
	return nil
}

func (m model) renderHelp() string {
	bindings := m.getKeyBindings()

	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))

	var parts []string
	for _, kb := range bindings {
		parts = append(parts, keyStyle.Render(kb.key)+descStyle.Render(":"+kb.desc))
	}
	return strings.Join(parts, " ")
}

func (m model) toQualityData() domain.QualityData {
	data := domain.QualityData{}

	for tagName, selected := range m.selectedTags {
		if selected {
			data.Tags = append(data.Tags, tagName)
		}
	}

	if m.promptSpecificity > 0 {
		v := m.promptSpecificity
		data.PromptSpecificity = &v
	}

	if m.taskCompletion > 0 {
		v := m.taskCompletion
		data.TaskCompletion = &v
	}

	if m.codeConfidence > 0 {
		v := m.codeConfidence
		data.CodeConfidence = &v
	}

	if m.rating > 0 {
		v := m.rating
		data.Rating = &v
	}

	data.Notes = strings.TrimSpace(m.notesInput.Value())

	return data
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
