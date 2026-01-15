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
	p.logger.Debug("CollectQualityData called")

	// Ensure TERM is set for proper terminal rendering
	if os.Getenv("TERM") == "" {
		os.Setenv("TERM", "xterm-256color")
		p.logger.Debug("TERM was empty, set to xterm-256color")
	}

	// Open /dev/tty directly since stdin is consumed by hook input
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		p.logger.Debug(fmt.Sprintf("TTY not available: %v, skipping quality prompts", err))
		return domain.QualityData{}, nil
	}
	defer tty.Close()

	return p.runTUI(tags, tty)
}

// CollectQualityDataWithTTY prompts using a pre-opened TTY file.
// Used by forked child processes that inherit TTY from parent.
func (p *BubbleTeaPrompter) CollectQualityDataWithTTY(tags []domain.Tag, tty *os.File) (domain.QualityData, error) {
	p.logger.Debug("CollectQualityDataWithTTY called")

	if os.Getenv("TERM") == "" {
		os.Setenv("TERM", "xterm-256color")
	}

	return p.runTUI(tags, tty)
}

func (p *BubbleTeaPrompter) runTUI(tags []domain.Tag, tty *os.File) (domain.QualityData, error) {
	p.logger.Debug(fmt.Sprintf("Starting TUI, TERM=%s", os.Getenv("TERM")))

	m := newModel(tags)
	prog := tea.NewProgram(m, tea.WithAltScreen(), tea.WithInput(tty), tea.WithOutput(tty))
	finalModel, err := prog.Run()
	if err != nil {
		p.logger.Debug(fmt.Sprintf("TUI error: %v", err))
		return domain.QualityData{}, err
	}

	result := finalModel.(model)
	if result.cancelled {
		p.logger.Debug("TUI cancelled by user")
		return domain.QualityData{}, nil
	}

	p.logger.Debug("TUI completed successfully")
	return result.toQualityData(), nil
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

// Monochrome grayscale styles with strong text hierarchy
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
	// Grayscale palette
	white := lipgloss.Color("#FFFFFF")
	black := lipgloss.Color("#000000")
	gray300 := lipgloss.Color("#E0E0E0")
	gray500 := lipgloss.Color("#9E9E9E")
	gray600 := lipgloss.Color("#757575")
	gray700 := lipgloss.Color("#616161")
	gray800 := lipgloss.Color("#424242")

	return styles{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(white),
		subtitle: lipgloss.NewStyle().
			Foreground(gray300).
			Bold(true),
		cursor: lipgloss.NewStyle().
			Foreground(black).
			Background(white).
			Bold(true),
		selected: lipgloss.NewStyle().
			Foreground(white).
			Bold(true),
		unselected: lipgloss.NewStyle().
			Foreground(gray600),
		help: lipgloss.NewStyle().
			Foreground(gray700).
			MarginTop(2),
		helpKey: lipgloss.NewStyle().
			Foreground(gray500).
			Bold(true),
		container: lipgloss.NewStyle().
			Padding(1, 2),
		indicator: lipgloss.NewStyle().
			Foreground(white).
			Bold(true),
		numberHint: lipgloss.NewStyle().
			Foreground(gray700),
		activeNum: lipgloss.NewStyle().
			Foreground(black).
			Background(white).
			Bold(true),
		progressBar: lipgloss.NewStyle().
			Foreground(gray800),
		progressDot: lipgloss.NewStyle().
			Foreground(white),
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
	case "b", "shift+tab":
		return m.prevStep()
	case "enter", "tab":
		m.step = stepDone
		return m, tea.Quit
	case "esc":
		// Skip and finish
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
		// Skip remaining and finish
		m.step = stepDone
		return m, tea.Quit

	// Back navigation - always go to previous step
	case "b", "shift+tab":
		return m.prevStep()

	// Forward navigation
	case "enter", "tab":
		return m.nextStep()

	// Vertical navigation (tags only)
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

	// Horizontal navigation (scales only)
	case "h", "left":
		if m.isScaleStep() && m.scaleCursor > 1 {
			m.scaleCursor--
		}
		return m, nil

	case "l", "right":
		if m.isScaleStep() && m.scaleCursor < 5 {
			m.scaleCursor++
		}
		return m, nil

	// Selection
	case " ":
		if m.step == stepTagsTaskType {
			m.toggleTag()
		} else if m.isScaleStep() {
			m.setCurrentScale(m.scaleCursor)
		}
		return m, nil

	// Quick number selection for scales
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

	// Header with uppercase title and progress
	b.WriteString(m.styles.title.Render("SESSION FEEDBACK"))
	b.WriteString("  ")
	b.WriteString(m.renderProgress())
	b.WriteString("\n")

	// Separator line
	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#404040")).
		Render("────────────────────────────────────────")
	b.WriteString(sep)
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

	// Clean step counter: [2/6]
	stepText := fmt.Sprintf("[%d/%d]", currentStep+1, totalSteps)
	counter := lipgloss.NewStyle().Foreground(lipgloss.Color("#737373")).Render(stepText)

	// Minimalist progress bar
	var bar strings.Builder
	for i := 0; i < totalSteps; i++ {
		if i <= currentStep {
			bar.WriteString(m.styles.progressDot.Render("━"))
		} else {
			bar.WriteString(m.styles.progressBar.Render("─"))
		}
	}

	return counter + " " + bar.String()
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

	b.WriteString(m.styles.subtitle.Render("TASK TYPE"))
	b.WriteString("\n\n")

	for i, tag := range m.taskTypeTags {
		isSelected := m.selectedTags[tag.Name]
		isCursor := i == m.tagCursor

		// Clean checkbox style
		var checkbox string
		if isSelected {
			checkbox = m.styles.selected.Render("[x]")
		} else {
			checkbox = m.styles.unselected.Render("[ ]")
		}

		// Tag name with cursor highlight (inverted)
		var name string
		if isCursor {
			name = m.styles.cursor.Render(" " + tag.Name + " ")
		} else if isSelected {
			name = m.styles.selected.Render(tag.Name)
		} else {
			name = m.styles.unselected.Render(tag.Name)
		}

		b.WriteString(fmt.Sprintf("  %s %s\n", checkbox, name))
	}

	return b.String()
}

func (m model) viewScale() string {
	var b strings.Builder

	info := scaleLabels[m.step]
	currentValue := m.getCurrentScaleValue()

	// Uppercase label
	b.WriteString(m.styles.subtitle.Render(strings.ToUpper(info.label)))
	b.WriteString("\n\n")

	// Scale description labels - spread evenly
	b.WriteString("  ")
	b.WriteString(m.styles.unselected.Render(info.lowDesc))
	// Calculate padding to right-align the high description
	padding := 35 - len(info.lowDesc) - len(info.highDesc)
	if padding < 4 {
		padding = 4
	}
	b.WriteString(strings.Repeat(" ", padding))
	b.WriteString(m.styles.unselected.Render(info.highDesc))
	b.WriteString("\n\n")

	// Number row - clean inverted selection
	b.WriteString("  ")
	for i := 1; i <= 5; i++ {
		isCursor := i == m.scaleCursor
		isSelected := i == currentValue
		numStr := fmt.Sprintf(" %d ", i)

		if isCursor {
			// Inverted: black on white background
			b.WriteString(m.styles.activeNum.Render(numStr))
		} else if isSelected {
			// Selected but not cursor: bold white
			b.WriteString(m.styles.selected.Render(numStr))
		} else {
			// Unselected: dim
			b.WriteString(m.styles.numberHint.Render(numStr))
		}
		b.WriteString("  ")
	}
	b.WriteString("\n")

	return b.String()
}

func (m model) viewNotes() string {
	var b strings.Builder

	b.WriteString(m.styles.subtitle.Render("NOTES"))
	b.WriteString("  ")
	b.WriteString(m.styles.unselected.Render("optional"))

	// Vim mode indicator - clean style
	b.WriteString("  ")
	if m.vimMode == modeInsert {
		b.WriteString(m.styles.activeNum.Render(" INSERT "))
	} else {
		b.WriteString(m.styles.unselected.Render("[NORMAL]"))
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
			{"j/k", "move"},
			{"spc", "toggle"},
			{"⏎", "next"},
			{"b", "back"},
			{"esc", "skip"},
		}
	case m.isScaleStep():
		return []keyBinding{
			{"h/l", "adjust"},
			{"1-5", "select"},
			{"⏎", "next"},
			{"b", "back"},
			{"esc", "skip"},
		}
	case m.step == stepNotes:
		if m.vimMode == modeInsert {
			return []keyBinding{
				{"esc", "exit edit"},
			}
		}
		return []keyBinding{
			{"i", "edit"},
			{"⏎", "done"},
			{"b", "back"},
			{"esc", "skip"},
		}
	}
	return nil
}

func (m model) renderHelp() string {
	bindings := m.getKeyBindings()

	// Monochrome help style
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#A3A3A3")).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#525252"))
	sepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#404040"))

	var parts []string
	for _, kb := range bindings {
		parts = append(parts, keyStyle.Render(kb.key)+descStyle.Render(":"+kb.desc))
	}
	return strings.Join(parts, sepStyle.Render("  /  "))
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
